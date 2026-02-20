use serde::{Deserialize, Serialize};
use std::fmt;

/// Post-quantum key encapsulation mechanism variants (FIPS 203).
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum MlKemVariant {
    MlKem512,
    MlKem768,
    MlKem1024,
}

/// Post-quantum digital signature variants (FIPS 204).
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum MlDsaVariant {
    MlDsa44,
    MlDsa65,
    MlDsa87,
}

/// Stateless hash-based digital signature variants (FIPS 205).
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum SlhDsaVariant {
    Sha2_128s,
    Sha2_128f,
    Sha2_192s,
    Sha2_192f,
    Sha2_256s,
    Sha2_256f,
}

/// Hybrid algorithms combining classical and post-quantum schemes.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum HybridVariant {
    X25519MlKem768,
    Ed25519MlDsa65,
}

/// Top-level algorithm enum covering all supported cryptographic algorithms.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Algorithm {
    MlKem(MlKemVariant),
    MlDsa(MlDsaVariant),
    SlhDsa(SlhDsaVariant),
    Hybrid(HybridVariant),
}

/// The type of a cryptographic key.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum KeyType {
    Kem,
    Signature,
    HybridKem,
    HybridSignature,
}

/// Permitted usages for a key.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum KeyUsage {
    Encrypt,
    Sign,
    KeyAgreement,
    Wrap,
}

impl Algorithm {
    /// Returns the key type implied by this algorithm.
    pub fn key_type(&self) -> KeyType {
        match self {
            Algorithm::MlKem(_) => KeyType::Kem,
            Algorithm::MlDsa(_) | Algorithm::SlhDsa(_) => KeyType::Signature,
            Algorithm::Hybrid(HybridVariant::X25519MlKem768) => KeyType::HybridKem,
            Algorithm::Hybrid(HybridVariant::Ed25519MlDsa65) => KeyType::HybridSignature,
        }
    }

    /// NIST security level (1 through 5).
    pub fn security_level(&self) -> u8 {
        match self {
            Algorithm::MlKem(MlKemVariant::MlKem512) => 1,
            Algorithm::MlKem(MlKemVariant::MlKem768) => 3,
            Algorithm::MlKem(MlKemVariant::MlKem1024) => 5,
            Algorithm::MlDsa(MlDsaVariant::MlDsa44) => 2,
            Algorithm::MlDsa(MlDsaVariant::MlDsa65) => 3,
            Algorithm::MlDsa(MlDsaVariant::MlDsa87) => 5,
            Algorithm::SlhDsa(v) => match v {
                SlhDsaVariant::Sha2_128s | SlhDsaVariant::Sha2_128f => 1,
                SlhDsaVariant::Sha2_192s | SlhDsaVariant::Sha2_192f => 3,
                SlhDsaVariant::Sha2_256s | SlhDsaVariant::Sha2_256f => 5,
            },
            Algorithm::Hybrid(HybridVariant::X25519MlKem768) => 3,
            Algorithm::Hybrid(HybridVariant::Ed25519MlDsa65) => 3,
        }
    }
}

impl fmt::Display for Algorithm {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Algorithm::MlKem(v) => write!(f, "{v}"),
            Algorithm::MlDsa(v) => write!(f, "{v}"),
            Algorithm::SlhDsa(v) => write!(f, "{v}"),
            Algorithm::Hybrid(v) => write!(f, "{v}"),
        }
    }
}

impl MlKemVariant {
    /// Returns (public_key_bytes, secret_key_bytes) per NIST spec.
    pub fn key_sizes(&self) -> (usize, usize) {
        match self {
            MlKemVariant::MlKem512 => (800, 1632),
            MlKemVariant::MlKem768 => (1184, 2400),
            MlKemVariant::MlKem1024 => (1568, 3168),
        }
    }

    /// Ciphertext size in bytes.
    pub fn ciphertext_size(&self) -> usize {
        match self {
            MlKemVariant::MlKem512 => 768,
            MlKemVariant::MlKem768 => 1088,
            MlKemVariant::MlKem1024 => 1568,
        }
    }
}

impl MlDsaVariant {
    /// Returns (public_key_bytes, secret_key_bytes) per NIST spec.
    pub fn key_sizes(&self) -> (usize, usize) {
        match self {
            MlDsaVariant::MlDsa44 => (1312, 2560),
            MlDsaVariant::MlDsa65 => (1952, 4032),
            MlDsaVariant::MlDsa87 => (2592, 4896),
        }
    }

    /// Signature size in bytes.
    pub fn signature_size(&self) -> usize {
        match self {
            MlDsaVariant::MlDsa44 => 2420,
            MlDsaVariant::MlDsa65 => 3309,
            MlDsaVariant::MlDsa87 => 4627,
        }
    }
}

impl SlhDsaVariant {
    /// Returns (public_key_bytes, secret_key_bytes) per NIST spec.
    pub fn key_sizes(&self) -> (usize, usize) {
        match self {
            SlhDsaVariant::Sha2_128s | SlhDsaVariant::Sha2_128f => (32, 64),
            SlhDsaVariant::Sha2_192s | SlhDsaVariant::Sha2_192f => (48, 96),
            SlhDsaVariant::Sha2_256s | SlhDsaVariant::Sha2_256f => (64, 128),
        }
    }

    /// Signature size in bytes. "s" variants are small/slow, "f" are fast/large.
    pub fn signature_size(&self) -> usize {
        match self {
            SlhDsaVariant::Sha2_128s => 7856,
            SlhDsaVariant::Sha2_128f => 17088,
            SlhDsaVariant::Sha2_192s => 16224,
            SlhDsaVariant::Sha2_192f => 35664,
            SlhDsaVariant::Sha2_256s => 29792,
            SlhDsaVariant::Sha2_256f => 49856,
        }
    }

    /// Whether this is the "small" (slower, smaller signature) variant.
    pub fn is_small(&self) -> bool {
        matches!(
            self,
            SlhDsaVariant::Sha2_128s | SlhDsaVariant::Sha2_192s | SlhDsaVariant::Sha2_256s
        )
    }
}

impl fmt::Display for MlKemVariant {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            MlKemVariant::MlKem512 => write!(f, "ML-KEM-512"),
            MlKemVariant::MlKem768 => write!(f, "ML-KEM-768"),
            MlKemVariant::MlKem1024 => write!(f, "ML-KEM-1024"),
        }
    }
}

impl fmt::Display for MlDsaVariant {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            MlDsaVariant::MlDsa44 => write!(f, "ML-DSA-44"),
            MlDsaVariant::MlDsa65 => write!(f, "ML-DSA-65"),
            MlDsaVariant::MlDsa87 => write!(f, "ML-DSA-87"),
        }
    }
}

impl fmt::Display for SlhDsaVariant {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            SlhDsaVariant::Sha2_128s => write!(f, "SLH-DSA-SHA2-128s"),
            SlhDsaVariant::Sha2_128f => write!(f, "SLH-DSA-SHA2-128f"),
            SlhDsaVariant::Sha2_192s => write!(f, "SLH-DSA-SHA2-192s"),
            SlhDsaVariant::Sha2_192f => write!(f, "SLH-DSA-SHA2-192f"),
            SlhDsaVariant::Sha2_256s => write!(f, "SLH-DSA-SHA2-256s"),
            SlhDsaVariant::Sha2_256f => write!(f, "SLH-DSA-SHA2-256f"),
        }
    }
}

impl fmt::Display for HybridVariant {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            HybridVariant::X25519MlKem768 => write!(f, "X25519-ML-KEM-768"),
            HybridVariant::Ed25519MlDsa65 => write!(f, "Ed25519-ML-DSA-65"),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn algorithm_security_levels() {
        assert_eq!(Algorithm::MlKem(MlKemVariant::MlKem512).security_level(), 1);
        assert_eq!(Algorithm::MlKem(MlKemVariant::MlKem768).security_level(), 3);
        assert_eq!(Algorithm::MlKem(MlKemVariant::MlKem1024).security_level(), 5);
        assert_eq!(Algorithm::MlDsa(MlDsaVariant::MlDsa87).security_level(), 5);
    }

    #[test]
    fn algorithm_key_types() {
        assert_eq!(
            Algorithm::MlKem(MlKemVariant::MlKem768).key_type(),
            KeyType::Kem
        );
        assert_eq!(
            Algorithm::MlDsa(MlDsaVariant::MlDsa65).key_type(),
            KeyType::Signature
        );
        assert_eq!(
            Algorithm::Hybrid(HybridVariant::X25519MlKem768).key_type(),
            KeyType::HybridKem
        );
    }

    #[test]
    fn algorithm_display() {
        assert_eq!(
            Algorithm::MlKem(MlKemVariant::MlKem768).to_string(),
            "ML-KEM-768"
        );
        assert_eq!(
            Algorithm::Hybrid(HybridVariant::X25519MlKem768).to_string(),
            "X25519-ML-KEM-768"
        );
    }
}
