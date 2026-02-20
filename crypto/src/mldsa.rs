use crate::error::{CryptoError, CryptoResult};
use ml_dsa::KeyGen;
use quantun_types::MlDsaVariant;
use serde::{Deserialize, Serialize};
use signature::{Signer, Verifier};
use zeroize::Zeroize;

/// ML-DSA key pair (FIPS 204).
///
/// Uses the `ml-dsa` crate (RustCrypto) for a standards-compliant
/// implementation of the Module-Lattice-Based Digital Signature Algorithm.
///
/// The secret key is stored as a 32-byte seed from which the full
/// signing key can be deterministically derived. The public key is
/// stored in its encoded form.
///
/// The secret key seed is automatically zeroized when dropped to prevent
/// key material from lingering in memory. It is excluded from serialization
/// to prevent accidental leakage via JSON/logging.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MlDsaKeyPair {
    pub variant: MlDsaVariant,
    /// Serialized verifying (public) key bytes.
    pub public_key: Vec<u8>,
    /// 32-byte seed from which the signing key is derived.
    /// Excluded from serialization to prevent accidental leakage.
    #[serde(skip)]
    pub secret_key: Vec<u8>,
}

impl Drop for MlDsaKeyPair {
    fn drop(&mut self) {
        self.secret_key.zeroize();
    }
}

/// An ML-DSA signature.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MlDsaSignature {
    pub signature: Vec<u8>,
    pub variant: MlDsaVariant,
}

impl MlDsaKeyPair {
    /// Generate a new ML-DSA key pair.
    pub fn generate(variant: MlDsaVariant) -> CryptoResult<Self> {
        // Generate random seed using OS CSPRNG (getrandom)
        let mut seed = [0u8; 32];
        getrandom::fill(&mut seed).expect("OS entropy source unavailable — cannot proceed safely");

        match variant {
            MlDsaVariant::MlDsa44 => {
                let kp = ml_dsa::MlDsa44::from_seed(&seed.into());
                Ok(make_keypair::<ml_dsa::MlDsa44>(variant, &kp))
            }
            MlDsaVariant::MlDsa65 => {
                let kp = ml_dsa::MlDsa65::from_seed(&seed.into());
                Ok(make_keypair::<ml_dsa::MlDsa65>(variant, &kp))
            }
            MlDsaVariant::MlDsa87 => {
                let kp = ml_dsa::MlDsa87::from_seed(&seed.into());
                Ok(make_keypair::<ml_dsa::MlDsa87>(variant, &kp))
            }
        }
    }

    /// Generate with a caller-supplied RNG. Delegates to OS RNG for PQC safety.
    pub fn generate_with_rng<R: rand::RngCore>(
        variant: MlDsaVariant,
        _rng: &mut R,
    ) -> CryptoResult<Self> {
        Self::generate(variant)
    }

    /// Sign a message.
    pub fn sign(&self, message: &[u8]) -> CryptoResult<MlDsaSignature> {
        match self.variant {
            MlDsaVariant::MlDsa44 => sign_impl::<ml_dsa::MlDsa44>(&self.secret_key, message, self.variant),
            MlDsaVariant::MlDsa65 => sign_impl::<ml_dsa::MlDsa65>(&self.secret_key, message, self.variant),
            MlDsaVariant::MlDsa87 => sign_impl::<ml_dsa::MlDsa87>(&self.secret_key, message, self.variant),
        }
    }

    /// Verify a signature against a message.
    pub fn verify(&self, message: &[u8], sig: &MlDsaSignature) -> CryptoResult<bool> {
        if sig.variant != self.variant {
            return Err(CryptoError::Verification(format!(
                "variant mismatch: key is {}, signature is {}",
                self.variant, sig.variant
            )));
        }

        match self.variant {
            MlDsaVariant::MlDsa44 => verify_impl::<ml_dsa::MlDsa44>(&self.public_key, message, &sig.signature),
            MlDsaVariant::MlDsa65 => verify_impl::<ml_dsa::MlDsa65>(&self.public_key, message, &sig.signature),
            MlDsaVariant::MlDsa87 => verify_impl::<ml_dsa::MlDsa87>(&self.public_key, message, &sig.signature),
        }
    }
}

/// Helper to build MlDsaKeyPair from a typed KeyPair.
fn make_keypair<P: ml_dsa::MlDsaParams>(
    variant: MlDsaVariant,
    kp: &ml_dsa::KeyPair<P>,
) -> MlDsaKeyPair {
    let public_key = kp.verifying_key().encode().to_vec();
    // Store the seed (32 bytes) for deterministic re-derivation
    let secret_key = kp.to_seed().to_vec();

    tracing::debug!(
        algorithm = %variant,
        pk_bytes = public_key.len(),
        sk_seed_bytes = secret_key.len(),
        "generated ML-DSA key pair (FIPS 204)"
    );

    MlDsaKeyPair {
        variant,
        public_key,
        secret_key,
    }
}

/// Sign a message using a serialized seed.
fn sign_impl<P: ml_dsa::MlDsaParams>(
    seed_bytes: &[u8],
    message: &[u8],
    variant: MlDsaVariant,
) -> CryptoResult<MlDsaSignature>
where
    ml_dsa::SigningKey<P>: Signer<ml_dsa::Signature<P>>,
{
    let mut seed: [u8; 32] = seed_bytes.try_into().map_err(|_| {
        CryptoError::Signing(format!(
            "invalid ML-DSA seed ({} bytes, expected 32)",
            seed_bytes.len()
        ))
    })?;
    let sk = ml_dsa::SigningKey::<P>::from_seed(&seed.into());
    // Zeroize the seed copy immediately after use
    seed.zeroize();
    let sig = sk.sign(message);
    let sig_bytes = sig.encode().to_vec();

    Ok(MlDsaSignature {
        signature: sig_bytes,
        variant,
    })
}

/// Verify a signature using serialized key and signature bytes.
fn verify_impl<P: ml_dsa::MlDsaParams>(
    vk_bytes: &[u8],
    message: &[u8],
    sig_bytes: &[u8],
) -> CryptoResult<bool>
where
    ml_dsa::VerifyingKey<P>: Verifier<ml_dsa::Signature<P>>,
{
    let vk_array = vk_bytes.try_into().map_err(|_| {
        CryptoError::Verification(format!(
            "invalid verifying key ({} bytes)",
            vk_bytes.len()
        ))
    })?;
    let vk = ml_dsa::VerifyingKey::<P>::decode(vk_array);

    let sig = ml_dsa::Signature::<P>::try_from(sig_bytes).map_err(|_| {
        CryptoError::Verification(format!(
            "invalid signature ({} bytes)",
            sig_bytes.len()
        ))
    })?;

    match vk.verify(message, &sig) {
        Ok(()) => Ok(true),
        Err(_) => Ok(false),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn keygen_produces_keys() {
        for variant in [
            MlDsaVariant::MlDsa44,
            MlDsaVariant::MlDsa65,
            MlDsaVariant::MlDsa87,
        ] {
            let kp = MlDsaKeyPair::generate(variant).unwrap();
            assert_eq!(kp.secret_key.len(), 32, "seed should be 32 bytes for {variant}");
            assert!(!kp.public_key.is_empty(), "public key must not be empty for {variant}");
        }
    }

    #[test]
    fn sign_verify_round_trip_44() {
        let kp = MlDsaKeyPair::generate(MlDsaVariant::MlDsa44).unwrap();
        let sig = kp.sign(b"hello quantum world").unwrap();
        assert!(kp.verify(b"hello quantum world", &sig).unwrap());
    }

    #[test]
    fn sign_verify_round_trip_65() {
        let kp = MlDsaKeyPair::generate(MlDsaVariant::MlDsa65).unwrap();
        let sig = kp.sign(b"hello quantum world").unwrap();
        assert!(kp.verify(b"hello quantum world", &sig).unwrap());
    }

    #[test]
    fn sign_verify_round_trip_87() {
        let kp = MlDsaKeyPair::generate(MlDsaVariant::MlDsa87).unwrap();
        let sig = kp.sign(b"hello quantum world").unwrap();
        assert!(kp.verify(b"hello quantum world", &sig).unwrap());
    }

    #[test]
    fn verify_wrong_message_fails() {
        let kp = MlDsaKeyPair::generate(MlDsaVariant::MlDsa44).unwrap();
        let sig = kp.sign(b"original").unwrap();
        assert!(!kp.verify(b"tampered", &sig).unwrap());
    }

    #[test]
    fn signature_correct_size() {
        let kp = MlDsaKeyPair::generate(MlDsaVariant::MlDsa87).unwrap();
        let sig = kp.sign(b"test").unwrap();
        assert_eq!(sig.signature.len(), MlDsaVariant::MlDsa87.signature_size());
    }

    #[test]
    fn variant_mismatch_errors() {
        let kp44 = MlDsaKeyPair::generate(MlDsaVariant::MlDsa44).unwrap();
        let sig65 = MlDsaSignature {
            signature: vec![0u8; MlDsaVariant::MlDsa65.signature_size()],
            variant: MlDsaVariant::MlDsa65,
        };
        assert!(kp44.verify(b"test", &sig65).is_err());
    }
}
