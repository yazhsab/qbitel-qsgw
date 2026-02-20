use serde::{Deserialize, Serialize};
use std::fmt;

/// Platform-wide error codes.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum ErrorCode {
    // Generic
    Internal,
    InvalidArgument,
    NotFound,
    AlreadyExists,
    PermissionDenied,
    Unauthenticated,

    // Crypto
    UnsupportedAlgorithm,
    KeyGenerationFailed,
    EncapsulationFailed,
    DecapsulationFailed,
    SigningFailed,
    VerificationFailed,
    InvalidKeyMaterial,
    KeyExpired,
    KeyRevoked,

    // TLS
    TlsHandshakeFailed,
    CertificateInvalid,
    CertificateExpired,

    // Device
    DeviceNotProvisioned,
    DeviceOffline,
    FirmwareIncompatible,

    // Risk
    AssessmentFailed,
    ScanTimeout,
}

impl ErrorCode {
    /// Returns a short string code suitable for API responses.
    pub fn as_str(&self) -> &'static str {
        match self {
            ErrorCode::Internal => "INTERNAL",
            ErrorCode::InvalidArgument => "INVALID_ARGUMENT",
            ErrorCode::NotFound => "NOT_FOUND",
            ErrorCode::AlreadyExists => "ALREADY_EXISTS",
            ErrorCode::PermissionDenied => "PERMISSION_DENIED",
            ErrorCode::Unauthenticated => "UNAUTHENTICATED",
            ErrorCode::UnsupportedAlgorithm => "UNSUPPORTED_ALGORITHM",
            ErrorCode::KeyGenerationFailed => "KEY_GENERATION_FAILED",
            ErrorCode::EncapsulationFailed => "ENCAPSULATION_FAILED",
            ErrorCode::DecapsulationFailed => "DECAPSULATION_FAILED",
            ErrorCode::SigningFailed => "SIGNING_FAILED",
            ErrorCode::VerificationFailed => "VERIFICATION_FAILED",
            ErrorCode::InvalidKeyMaterial => "INVALID_KEY_MATERIAL",
            ErrorCode::KeyExpired => "KEY_EXPIRED",
            ErrorCode::KeyRevoked => "KEY_REVOKED",
            ErrorCode::TlsHandshakeFailed => "TLS_HANDSHAKE_FAILED",
            ErrorCode::CertificateInvalid => "CERTIFICATE_INVALID",
            ErrorCode::CertificateExpired => "CERTIFICATE_EXPIRED",
            ErrorCode::DeviceNotProvisioned => "DEVICE_NOT_PROVISIONED",
            ErrorCode::DeviceOffline => "DEVICE_OFFLINE",
            ErrorCode::FirmwareIncompatible => "FIRMWARE_INCOMPATIBLE",
            ErrorCode::AssessmentFailed => "ASSESSMENT_FAILED",
            ErrorCode::ScanTimeout => "SCAN_TIMEOUT",
        }
    }
}

impl fmt::Display for ErrorCode {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn error_code_display() {
        assert_eq!(ErrorCode::Internal.to_string(), "INTERNAL");
        assert_eq!(
            ErrorCode::UnsupportedAlgorithm.to_string(),
            "UNSUPPORTED_ALGORITHM"
        );
    }

    #[test]
    fn error_code_as_str() {
        assert_eq!(ErrorCode::KeyExpired.as_str(), "KEY_EXPIRED");
    }
}
