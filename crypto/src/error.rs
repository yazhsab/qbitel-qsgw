use quantun_types::ErrorCode;
use thiserror::Error;

/// Errors produced by cryptographic operations.
#[derive(Debug, Error)]
pub enum CryptoError {
    #[error("key generation failed for {algorithm}: {reason}")]
    KeyGeneration { algorithm: String, reason: String },

    #[error("encapsulation failed: {0}")]
    Encapsulation(String),

    #[error("decapsulation failed: {0}")]
    Decapsulation(String),

    #[error("signing failed: {0}")]
    Signing(String),

    #[error("verification failed: {0}")]
    Verification(String),

    #[error("invalid key material: {0}")]
    InvalidKeyMaterial(String),

    #[error("unsupported algorithm: {0}")]
    UnsupportedAlgorithm(String),

    #[error("serialization error: {0}")]
    Serialization(String),

    #[error("rng error: {0}")]
    Rng(String),
}

impl CryptoError {
    /// Map to a platform error code.
    pub fn error_code(&self) -> ErrorCode {
        match self {
            CryptoError::KeyGeneration { .. } => ErrorCode::KeyGenerationFailed,
            CryptoError::Encapsulation(_) => ErrorCode::EncapsulationFailed,
            CryptoError::Decapsulation(_) => ErrorCode::DecapsulationFailed,
            CryptoError::Signing(_) => ErrorCode::SigningFailed,
            CryptoError::Verification(_) => ErrorCode::VerificationFailed,
            CryptoError::InvalidKeyMaterial(_) => ErrorCode::InvalidKeyMaterial,
            CryptoError::UnsupportedAlgorithm(_) => ErrorCode::UnsupportedAlgorithm,
            CryptoError::Serialization(_) => ErrorCode::Internal,
            CryptoError::Rng(_) => ErrorCode::Internal,
        }
    }
}

pub type CryptoResult<T> = Result<T, CryptoError>;
