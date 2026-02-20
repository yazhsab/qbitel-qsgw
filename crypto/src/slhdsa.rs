use crate::error::{CryptoError, CryptoResult};
use quantun_types::SlhDsaVariant;
use serde::{Deserialize, Serialize};
use signature::Verifier;
use zeroize::Zeroize;

/// SLH-DSA key pair (FIPS 205).
///
/// Uses the `slh-dsa` crate (RustCrypto) for a standards-compliant
/// implementation of the Stateless Hash-Based Digital Signature Algorithm.
/// SLH-DSA provides conservative post-quantum security based solely on
/// the security of hash functions, making it a useful fallback when
/// lattice-based schemes face new cryptanalysis.
///
/// The secret key is automatically zeroized when dropped to prevent key
/// material from lingering in memory. It is excluded from serialization
/// to prevent accidental leakage via JSON/logging.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SlhDsaKeyPair {
    pub variant: SlhDsaVariant,
    /// Serialized verifying (public) key bytes.
    pub public_key: Vec<u8>,
    /// Serialized signing (secret) key bytes.
    /// Excluded from serialization to prevent accidental leakage.
    #[serde(skip)]
    pub secret_key: Vec<u8>,
}

impl Drop for SlhDsaKeyPair {
    fn drop(&mut self) {
        self.secret_key.zeroize();
    }
}

/// An SLH-DSA signature.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SlhDsaSignature {
    pub signature: Vec<u8>,
    pub variant: SlhDsaVariant,
}

impl SlhDsaKeyPair {
    /// Generate a new SLH-DSA key pair using OS RNG.
    pub fn generate(variant: SlhDsaVariant) -> CryptoResult<Self> {
        match variant {
            SlhDsaVariant::Sha2_128s => generate_typed::<slh_dsa::Sha2_128s>(variant),
            SlhDsaVariant::Sha2_128f => generate_typed::<slh_dsa::Sha2_128f>(variant),
            SlhDsaVariant::Sha2_192s => generate_typed::<slh_dsa::Sha2_192s>(variant),
            SlhDsaVariant::Sha2_192f => generate_typed::<slh_dsa::Sha2_192f>(variant),
            SlhDsaVariant::Sha2_256s => generate_typed::<slh_dsa::Sha2_256s>(variant),
            SlhDsaVariant::Sha2_256f => generate_typed::<slh_dsa::Sha2_256f>(variant),
        }
    }

    /// Generate with a caller-supplied RNG. Delegates to OS RNG for PQC safety.
    pub fn generate_with_rng<R: rand::RngCore>(
        variant: SlhDsaVariant,
        _rng: &mut R,
    ) -> CryptoResult<Self> {
        Self::generate(variant)
    }

    /// Sign a message using OS RNG for randomized signing.
    pub fn sign(&self, message: &[u8]) -> CryptoResult<SlhDsaSignature> {
        match self.variant {
            SlhDsaVariant::Sha2_128s => sign_typed::<slh_dsa::Sha2_128s>(&self.secret_key, message, self.variant),
            SlhDsaVariant::Sha2_128f => sign_typed::<slh_dsa::Sha2_128f>(&self.secret_key, message, self.variant),
            SlhDsaVariant::Sha2_192s => sign_typed::<slh_dsa::Sha2_192s>(&self.secret_key, message, self.variant),
            SlhDsaVariant::Sha2_192f => sign_typed::<slh_dsa::Sha2_192f>(&self.secret_key, message, self.variant),
            SlhDsaVariant::Sha2_256s => sign_typed::<slh_dsa::Sha2_256s>(&self.secret_key, message, self.variant),
            SlhDsaVariant::Sha2_256f => sign_typed::<slh_dsa::Sha2_256f>(&self.secret_key, message, self.variant),
        }
    }

    /// Verify a signature against a message.
    pub fn verify(&self, message: &[u8], sig: &SlhDsaSignature) -> CryptoResult<bool> {
        if sig.variant != self.variant {
            return Err(CryptoError::Verification(format!(
                "variant mismatch: key is {}, signature is {}",
                self.variant, sig.variant
            )));
        }

        match self.variant {
            SlhDsaVariant::Sha2_128s => verify_typed::<slh_dsa::Sha2_128s>(&self.public_key, message, &sig.signature),
            SlhDsaVariant::Sha2_128f => verify_typed::<slh_dsa::Sha2_128f>(&self.public_key, message, &sig.signature),
            SlhDsaVariant::Sha2_192s => verify_typed::<slh_dsa::Sha2_192s>(&self.public_key, message, &sig.signature),
            SlhDsaVariant::Sha2_192f => verify_typed::<slh_dsa::Sha2_192f>(&self.public_key, message, &sig.signature),
            SlhDsaVariant::Sha2_256s => verify_typed::<slh_dsa::Sha2_256s>(&self.public_key, message, &sig.signature),
            SlhDsaVariant::Sha2_256f => verify_typed::<slh_dsa::Sha2_256f>(&self.public_key, message, &sig.signature),
        }
    }
}

/// Generate a key pair for a concrete SLH-DSA parameter set.
fn generate_typed<P>(variant: SlhDsaVariant) -> CryptoResult<SlhDsaKeyPair>
where
    P: slh_dsa::ParameterSet,
{
    let sk = slh_dsa::SigningKey::<P>::new(&mut crate::rng::PqcRng);
    let vk: &slh_dsa::VerifyingKey<P> = sk.as_ref();

    let secret_key = sk.to_vec();
    let public_key = vk.to_vec();

    tracing::debug!(
        algorithm = %variant,
        pk_bytes = public_key.len(),
        sk_bytes = secret_key.len(),
        "generated SLH-DSA key pair (FIPS 205)"
    );

    Ok(SlhDsaKeyPair {
        variant,
        public_key,
        secret_key,
    })
}

/// Sign a message with a serialized signing key.
fn sign_typed<P>(
    sk_bytes: &[u8],
    message: &[u8],
    variant: SlhDsaVariant,
) -> CryptoResult<SlhDsaSignature>
where
    P: slh_dsa::ParameterSet,
{
    use signature::RandomizedSigner;

    let sk = slh_dsa::SigningKey::<P>::try_from(sk_bytes).map_err(|_| {
        CryptoError::Signing(format!(
            "invalid SLH-DSA signing key ({} bytes)",
            sk_bytes.len()
        ))
    })?;

    let sig = sk.try_sign_with_rng(&mut crate::rng::PqcRng, message).map_err(|e| {
        CryptoError::Signing(format!("SLH-DSA signing failed: {e}"))
    })?;
    let sig_bytes = sig.to_vec();

    Ok(SlhDsaSignature {
        signature: sig_bytes,
        variant,
    })
}

/// Verify a signature with a serialized verifying key.
fn verify_typed<P>(
    vk_bytes: &[u8],
    message: &[u8],
    sig_bytes: &[u8],
) -> CryptoResult<bool>
where
    P: slh_dsa::ParameterSet,
{
    let vk = slh_dsa::VerifyingKey::<P>::try_from(vk_bytes).map_err(|_| {
        CryptoError::Verification(format!(
            "invalid SLH-DSA verifying key ({} bytes)",
            vk_bytes.len()
        ))
    })?;

    let sig = slh_dsa::Signature::<P>::try_from(sig_bytes).map_err(|_| {
        CryptoError::Verification(format!(
            "invalid SLH-DSA signature ({} bytes)",
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
    fn keygen_correct_sizes() {
        let kp = SlhDsaKeyPair::generate(SlhDsaVariant::Sha2_128s).unwrap();
        let (pk_len, sk_len) = SlhDsaVariant::Sha2_128s.key_sizes();
        assert_eq!(kp.public_key.len(), pk_len);
        assert_eq!(kp.secret_key.len(), sk_len);
    }

    #[test]
    fn sign_verify_round_trip_128s() {
        let kp = SlhDsaKeyPair::generate(SlhDsaVariant::Sha2_128s).unwrap();
        let sig = kp.sign(b"test message").unwrap();
        assert!(kp.verify(b"test message", &sig).unwrap());
    }

    #[test]
    fn sign_verify_round_trip_128f() {
        let kp = SlhDsaKeyPair::generate(SlhDsaVariant::Sha2_128f).unwrap();
        let sig = kp.sign(b"fast variant test").unwrap();
        assert!(kp.verify(b"fast variant test", &sig).unwrap());
    }

    #[test]
    fn verify_wrong_message() {
        let kp = SlhDsaKeyPair::generate(SlhDsaVariant::Sha2_128s).unwrap();
        let sig = kp.sign(b"correct").unwrap();
        assert!(!kp.verify(b"wrong", &sig).unwrap());
    }

    #[test]
    fn signature_correct_size() {
        let kp = SlhDsaKeyPair::generate(SlhDsaVariant::Sha2_128s).unwrap();
        let sig = kp.sign(b"test").unwrap();
        assert_eq!(sig.signature.len(), SlhDsaVariant::Sha2_128s.signature_size());
    }

    #[test]
    fn is_small_variant() {
        assert!(SlhDsaVariant::Sha2_128s.is_small());
        assert!(!SlhDsaVariant::Sha2_128f.is_small());
    }

    #[test]
    fn variant_mismatch_errors() {
        let kp = SlhDsaKeyPair::generate(SlhDsaVariant::Sha2_128s).unwrap();
        let wrong_sig = SlhDsaSignature {
            signature: vec![0u8; SlhDsaVariant::Sha2_128f.signature_size()],
            variant: SlhDsaVariant::Sha2_128f,
        };
        assert!(kp.verify(b"test", &wrong_sig).is_err());
    }
}
