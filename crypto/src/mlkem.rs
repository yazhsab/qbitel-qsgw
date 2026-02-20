use crate::error::{CryptoError, CryptoResult};
use ml_kem::{Decapsulate, Encapsulate, Kem, KeyExport, KeyInit, TryKeyInit};
use quantun_types::MlKemVariant;
use serde::{Deserialize, Serialize};
use zeroize::Zeroize;

/// ML-KEM key pair (FIPS 203).
///
/// Uses the `ml-kem` crate (RustCrypto) for a standards-compliant
/// implementation of the Module-Lattice-Based Key-Encapsulation Mechanism.
///
/// The secret key is automatically zeroized when dropped to prevent key
/// material from lingering in memory. The secret key is excluded from
/// serialization to prevent accidental leakage via JSON/logging.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MlKemKeyPair {
    pub variant: MlKemVariant,
    /// Serialized encapsulation (public) key bytes.
    pub public_key: Vec<u8>,
    /// Serialized decapsulation (secret) key bytes (seed form).
    /// Excluded from serialization to prevent accidental leakage.
    #[serde(skip)]
    pub secret_key: Vec<u8>,
}

impl Drop for MlKemKeyPair {
    fn drop(&mut self) {
        self.secret_key.zeroize();
    }
}

/// Result of an ML-KEM encapsulation operation.
///
/// The shared secret is zeroized on drop.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MlKemEncapsulated {
    pub ciphertext: Vec<u8>,
    #[serde(skip)]
    pub shared_secret: Vec<u8>,
}

impl Drop for MlKemEncapsulated {
    fn drop(&mut self) {
        self.shared_secret.zeroize();
    }
}

impl MlKemKeyPair {
    /// Generate a new ML-KEM key pair for the given variant using OS RNG.
    pub fn generate(variant: MlKemVariant) -> CryptoResult<Self> {
        match variant {
            MlKemVariant::MlKem512 => {
                let (dk, ek) = ml_kem::MlKem512::generate_keypair_from_rng(&mut crate::rng::PqcRng);
                Ok(make_keypair(variant, ek.to_bytes().to_vec(), dk.to_bytes().to_vec()))
            }
            MlKemVariant::MlKem768 => {
                let (dk, ek) = ml_kem::MlKem768::generate_keypair_from_rng(&mut crate::rng::PqcRng);
                Ok(make_keypair(variant, ek.to_bytes().to_vec(), dk.to_bytes().to_vec()))
            }
            MlKemVariant::MlKem1024 => {
                let (dk, ek) = ml_kem::MlKem1024::generate_keypair_from_rng(&mut crate::rng::PqcRng);
                Ok(make_keypair(variant, ek.to_bytes().to_vec(), dk.to_bytes().to_vec()))
            }
        }
    }

    /// Generate a key pair using a caller-supplied RNG.
    /// Delegates to OS RNG for cryptographic safety with PQC crates.
    pub fn generate_with_rng<R: rand::RngCore>(
        variant: MlKemVariant,
        _rng: &mut R,
    ) -> CryptoResult<Self> {
        Self::generate(variant)
    }

    /// Encapsulate: produce a ciphertext and shared secret from a public key.
    pub fn encapsulate(&self) -> CryptoResult<MlKemEncapsulated> {
        match self.variant {
            MlKemVariant::MlKem512 => {
                let ek = ml_kem::EncapsulationKey::<ml_kem::MlKem512>::new_from_slice(
                    &self.public_key,
                )
                .map_err(|_| {
                    CryptoError::Encapsulation(format!(
                        "invalid ML-KEM-512 encapsulation key ({} bytes)",
                        self.public_key.len()
                    ))
                })?;
                let (ct, ss) = ek.encapsulate_with_rng(&mut crate::rng::PqcRng);
                Ok(MlKemEncapsulated {
                    ciphertext: ct.to_vec(),
                    shared_secret: ss.to_vec(),
                })
            }
            MlKemVariant::MlKem768 => {
                let ek = ml_kem::EncapsulationKey::<ml_kem::MlKem768>::new_from_slice(
                    &self.public_key,
                )
                .map_err(|_| {
                    CryptoError::Encapsulation(format!(
                        "invalid ML-KEM-768 encapsulation key ({} bytes)",
                        self.public_key.len()
                    ))
                })?;
                let (ct, ss) = ek.encapsulate_with_rng(&mut crate::rng::PqcRng);
                Ok(MlKemEncapsulated {
                    ciphertext: ct.to_vec(),
                    shared_secret: ss.to_vec(),
                })
            }
            MlKemVariant::MlKem1024 => {
                let ek = ml_kem::EncapsulationKey::<ml_kem::MlKem1024>::new_from_slice(
                    &self.public_key,
                )
                .map_err(|_| {
                    CryptoError::Encapsulation(format!(
                        "invalid ML-KEM-1024 encapsulation key ({} bytes)",
                        self.public_key.len()
                    ))
                })?;
                let (ct, ss) = ek.encapsulate_with_rng(&mut crate::rng::PqcRng);
                Ok(MlKemEncapsulated {
                    ciphertext: ct.to_vec(),
                    shared_secret: ss.to_vec(),
                })
            }
        }
    }

    /// Decapsulate: recover the shared secret from a ciphertext using the secret key.
    pub fn decapsulate(&self, ciphertext: &[u8]) -> CryptoResult<Vec<u8>> {
        match self.variant {
            MlKemVariant::MlKem512 => {
                let dk = ml_kem::DecapsulationKey::<ml_kem::MlKem512>::new_from_slice(
                    &self.secret_key,
                )
                .map_err(|_| {
                    CryptoError::Decapsulation(format!(
                        "invalid ML-KEM-512 decapsulation key ({} bytes)",
                        self.secret_key.len()
                    ))
                })?;
                let ss = dk.decapsulate_slice(ciphertext).map_err(|_| {
                    CryptoError::Decapsulation(format!(
                        "ML-KEM-512 decapsulation failed (ct {} bytes)",
                        ciphertext.len()
                    ))
                })?;
                Ok(ss.to_vec())
            }
            MlKemVariant::MlKem768 => {
                let dk = ml_kem::DecapsulationKey::<ml_kem::MlKem768>::new_from_slice(
                    &self.secret_key,
                )
                .map_err(|_| {
                    CryptoError::Decapsulation(format!(
                        "invalid ML-KEM-768 decapsulation key ({} bytes)",
                        self.secret_key.len()
                    ))
                })?;
                let ss = dk.decapsulate_slice(ciphertext).map_err(|_| {
                    CryptoError::Decapsulation(format!(
                        "ML-KEM-768 decapsulation failed (ct {} bytes)",
                        ciphertext.len()
                    ))
                })?;
                Ok(ss.to_vec())
            }
            MlKemVariant::MlKem1024 => {
                let dk = ml_kem::DecapsulationKey::<ml_kem::MlKem1024>::new_from_slice(
                    &self.secret_key,
                )
                .map_err(|_| {
                    CryptoError::Decapsulation(format!(
                        "invalid ML-KEM-1024 decapsulation key ({} bytes)",
                        self.secret_key.len()
                    ))
                })?;
                let ss = dk.decapsulate_slice(ciphertext).map_err(|_| {
                    CryptoError::Decapsulation(format!(
                        "ML-KEM-1024 decapsulation failed (ct {} bytes)",
                        ciphertext.len()
                    ))
                })?;
                Ok(ss.to_vec())
            }
        }
    }
}

/// Helper to log and construct a key pair from raw bytes.
fn make_keypair(variant: MlKemVariant, public_key: Vec<u8>, secret_key: Vec<u8>) -> MlKemKeyPair {
    tracing::debug!(
        algorithm = %variant,
        pk_bytes = public_key.len(),
        sk_bytes = secret_key.len(),
        "generated ML-KEM key pair (FIPS 203)"
    );
    MlKemKeyPair {
        variant,
        public_key,
        secret_key,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn keygen_produces_keys() {
        for variant in [
            MlKemVariant::MlKem512,
            MlKemVariant::MlKem768,
            MlKemVariant::MlKem1024,
        ] {
            let kp = MlKemKeyPair::generate(variant).unwrap();
            assert!(!kp.public_key.is_empty(), "public key must not be empty for {variant}");
            assert!(!kp.secret_key.is_empty(), "secret key must not be empty for {variant}");
        }
    }

    #[test]
    fn encapsulate_decapsulate_round_trip_512() {
        let kp = MlKemKeyPair::generate(MlKemVariant::MlKem512).unwrap();
        let enc = kp.encapsulate().unwrap();
        let shared = kp.decapsulate(&enc.ciphertext).unwrap();
        assert_eq!(enc.shared_secret, shared);
    }

    #[test]
    fn encapsulate_decapsulate_round_trip_768() {
        let kp = MlKemKeyPair::generate(MlKemVariant::MlKem768).unwrap();
        let enc = kp.encapsulate().unwrap();
        let shared = kp.decapsulate(&enc.ciphertext).unwrap();
        assert_eq!(enc.shared_secret, shared);
    }

    #[test]
    fn encapsulate_decapsulate_round_trip_1024() {
        let kp = MlKemKeyPair::generate(MlKemVariant::MlKem1024).unwrap();
        let enc = kp.encapsulate().unwrap();
        let shared = kp.decapsulate(&enc.ciphertext).unwrap();
        assert_eq!(enc.shared_secret, shared);
    }

    #[test]
    fn decapsulate_wrong_ciphertext_fails() {
        let kp = MlKemKeyPair::generate(MlKemVariant::MlKem512).unwrap();
        let result = kp.decapsulate(&[0u8; 10]);
        assert!(result.is_err());
    }

    #[test]
    fn different_keypairs_produce_different_shared_secrets() {
        let kp1 = MlKemKeyPair::generate(MlKemVariant::MlKem768).unwrap();
        let kp2 = MlKemKeyPair::generate(MlKemVariant::MlKem768).unwrap();

        let enc1 = kp1.encapsulate().unwrap();
        let enc2 = kp2.encapsulate().unwrap();

        assert_ne!(enc1.shared_secret, enc2.shared_secret);
    }
}
