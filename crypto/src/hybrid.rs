use crate::error::{CryptoError, CryptoResult};
use crate::mlkem::MlKemKeyPair;
use quantun_types::{HybridVariant, MlKemVariant};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use x25519_dalek::{PublicKey, StaticSecret};
use zeroize::Zeroize;

/// Hybrid KEM key pair combining X25519 with ML-KEM-768.
///
/// Provides security against both classical and quantum adversaries by
/// combining the shared secrets from both schemes using a KDF.
/// This follows the hybrid approach recommended by NIST for the
/// transition period to post-quantum cryptography.
///
/// The classical secret key is automatically zeroized when dropped.
/// Secret key material is excluded from serialization to prevent
/// accidental leakage via JSON/logging.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HybridKemKeyPair {
    pub variant: HybridVariant,
    pub classical_public: Vec<u8>,
    #[serde(skip)]
    pub classical_secret: Option<Vec<u8>>,
    pub pqc_keypair: MlKemKeyPair,
}

impl Drop for HybridKemKeyPair {
    fn drop(&mut self) {
        if let Some(ref mut secret) = self.classical_secret {
            secret.zeroize();
        }
    }
}

/// Result of a hybrid encapsulation.
///
/// The shared secret is zeroized on drop.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HybridEncapsulated {
    pub classical_public: Vec<u8>,
    pub pqc_ciphertext: Vec<u8>,
    #[serde(skip)]
    pub shared_secret: Vec<u8>,
}

impl Drop for HybridEncapsulated {
    fn drop(&mut self) {
        self.shared_secret.zeroize();
    }
}

impl HybridKemKeyPair {
    /// Generate a new X25519 + ML-KEM-768 hybrid key pair.
    pub fn generate() -> CryptoResult<Self> {
        // Generate X25519 key pair using OS CSPRNG (getrandom)
        let mut key_bytes = [0u8; 32];
        getrandom::fill(&mut key_bytes)
            .expect("OS entropy source unavailable — cannot proceed safely");
        let classical_secret = StaticSecret::from(key_bytes);
        let classical_public = PublicKey::from(&classical_secret);

        // Generate ML-KEM-768 key pair using real FIPS 203 (uses OS RNG internally)
        let pqc_keypair = MlKemKeyPair::generate(MlKemVariant::MlKem768)?;

        let result = Self {
            variant: HybridVariant::X25519MlKem768,
            classical_public: classical_public.as_bytes().to_vec(),
            classical_secret: Some(key_bytes.to_vec()),
            pqc_keypair,
        };

        // Zeroize the stack copy of key_bytes
        key_bytes.zeroize();

        Ok(result)
    }

    /// Encapsulate against this key pair's public components.
    pub fn encapsulate(&self) -> CryptoResult<HybridEncapsulated> {
        // X25519 ephemeral key exchange using OS CSPRNG
        let mut ephemeral_bytes = [0u8; 32];
        getrandom::fill(&mut ephemeral_bytes)
            .expect("OS entropy source unavailable — cannot proceed safely");
        let ephemeral_secret = StaticSecret::from(ephemeral_bytes);
        let ephemeral_public = PublicKey::from(&ephemeral_secret);

        // Zeroize ephemeral bytes on stack
        ephemeral_bytes.zeroize();

        let recipient_public = PublicKey::from(
            <[u8; 32]>::try_from(self.classical_public.as_slice()).map_err(|_| {
                CryptoError::InvalidKeyMaterial("X25519 public key must be 32 bytes".into())
            })?,
        );
        let classical_shared = ephemeral_secret.diffie_hellman(&recipient_public);

        // Real ML-KEM-768 encapsulation (FIPS 203)
        let mut pqc_enc = self.pqc_keypair.encapsulate()?;

        // Combine both shared secrets via KDF
        let shared_secret =
            combine_secrets(classical_shared.as_bytes(), &pqc_enc.shared_secret);

        // Take ownership of ciphertext without moving out of Drop type
        let pqc_ciphertext = std::mem::take(&mut pqc_enc.ciphertext);

        Ok(HybridEncapsulated {
            classical_public: ephemeral_public.as_bytes().to_vec(),
            pqc_ciphertext,
            shared_secret,
        })
    }

    /// Decapsulate from ciphertext components.
    pub fn decapsulate(
        &self,
        ephemeral_public_bytes: &[u8],
        pqc_ciphertext: &[u8],
    ) -> CryptoResult<Vec<u8>> {
        let secret_bytes = self
            .classical_secret
            .as_ref()
            .ok_or_else(|| {
                CryptoError::InvalidKeyMaterial("secret key not available".into())
            })?;

        let secret_array: [u8; 32] = secret_bytes.as_slice().try_into().map_err(|_| {
            CryptoError::InvalidKeyMaterial("X25519 secret must be 32 bytes".into())
        })?;

        let classical_secret = StaticSecret::from(secret_array);

        let ephemeral_public = PublicKey::from(
            <[u8; 32]>::try_from(ephemeral_public_bytes).map_err(|_| {
                CryptoError::InvalidKeyMaterial(
                    "ephemeral public key must be 32 bytes".into(),
                )
            })?,
        );

        // X25519 shared secret
        let classical_shared = classical_secret.diffie_hellman(&ephemeral_public);

        // Real ML-KEM-768 decapsulation (FIPS 203)
        let pqc_shared = self.pqc_keypair.decapsulate(pqc_ciphertext)?;

        // Combine both shared secrets via KDF
        let shared_secret =
            combine_secrets(classical_shared.as_bytes(), &pqc_shared);

        Ok(shared_secret)
    }
}

/// KDF: combine classical and PQC shared secrets.
///
/// Uses SHA-256 with a domain separator to derive the final shared secret.
/// This ensures that the combined key is at least as strong as the stronger
/// of the two component schemes.
fn combine_secrets(classical: &[u8], pqc: &[u8]) -> Vec<u8> {
    let mut hasher = Sha256::new();
    hasher.update(b"quantun-hybrid-kem-v1");
    hasher.update(classical);
    hasher.update(pqc);
    hasher.finalize().to_vec()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn hybrid_keygen() {
        let kp = HybridKemKeyPair::generate().unwrap();
        assert_eq!(kp.classical_public.len(), 32);
        assert_eq!(kp.variant, HybridVariant::X25519MlKem768);
        assert!(!kp.pqc_keypair.public_key.is_empty());
        assert!(!kp.pqc_keypair.secret_key.is_empty());
    }

    #[test]
    fn hybrid_encapsulate_decapsulate() {
        let kp = HybridKemKeyPair::generate().unwrap();
        let enc = kp.encapsulate().unwrap();
        let shared = kp
            .decapsulate(&enc.classical_public, &enc.pqc_ciphertext)
            .unwrap();
        assert_eq!(enc.shared_secret, shared);
        assert_eq!(shared.len(), 32); // SHA-256 output
    }

    #[test]
    fn different_encapsulations_produce_different_secrets() {
        let kp = HybridKemKeyPair::generate().unwrap();
        let enc1 = kp.encapsulate().unwrap();
        let enc2 = kp.encapsulate().unwrap();
        // Each encapsulation uses fresh ephemeral keys
        assert_ne!(enc1.shared_secret, enc2.shared_secret);
    }

    #[test]
    fn missing_secret_key_errors() {
        let mut kp = HybridKemKeyPair::generate().unwrap();
        kp.classical_secret = None;
        let enc = kp.encapsulate().unwrap();
        assert!(kp
            .decapsulate(&enc.classical_public, &enc.pqc_ciphertext)
            .is_err());
    }
}
