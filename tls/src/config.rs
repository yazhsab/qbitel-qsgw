use quantun_types::Algorithm;
use serde::{Deserialize, Serialize};
use std::path::PathBuf;

/// TLS configuration for quantum-safe connections.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TlsConfig {
    /// Path to the PEM-encoded certificate chain.
    pub cert_path: PathBuf,
    /// Path to the PEM-encoded private key.
    pub key_path: PathBuf,
    /// Optional path to a custom CA bundle for verification.
    pub ca_path: Option<PathBuf>,
    /// Preferred post-quantum algorithms for key exchange, in priority order.
    pub preferred_algorithms: Vec<Algorithm>,
    /// Minimum TLS version (defaults to 1.3).
    pub min_tls_version: TlsVersion,
    /// Whether to require mutual TLS (client certificates).
    pub mutual_tls: bool,
    /// Whether to enable hybrid key exchange (classical + PQC).
    pub hybrid_mode: bool,
}

/// Supported TLS protocol versions.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TlsVersion {
    Tls12,
    Tls13,
}

/// Quantum-safe cipher suite identifier.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PqcCipherSuite {
    /// TLS_AES_256_GCM_SHA384 with X25519+ML-KEM-768 key exchange.
    Aes256GcmX25519MlKem768,
    /// TLS_AES_128_GCM_SHA256 with X25519+ML-KEM-512 key exchange.
    Aes128GcmX25519MlKem512,
    /// TLS_AES_256_GCM_SHA384 with ML-KEM-1024 key exchange.
    Aes256GcmMlKem1024,
}

impl Default for TlsConfig {
    fn default() -> Self {
        Self {
            cert_path: PathBuf::from("certs/server.pem"),
            key_path: PathBuf::from("certs/server-key.pem"),
            ca_path: None,
            preferred_algorithms: vec![
                Algorithm::Hybrid(quantun_types::HybridVariant::X25519MlKem768),
                Algorithm::MlKem(quantun_types::MlKemVariant::MlKem768),
            ],
            min_tls_version: TlsVersion::Tls13,
            mutual_tls: false,
            hybrid_mode: true,
        }
    }
}

impl TlsConfig {
    /// Create a config for development/testing with self-signed certs.
    pub fn development() -> Self {
        Self {
            cert_path: PathBuf::from("certs/dev.pem"),
            key_path: PathBuf::from("certs/dev-key.pem"),
            ca_path: None,
            preferred_algorithms: vec![Algorithm::Hybrid(
                quantun_types::HybridVariant::X25519MlKem768,
            )],
            min_tls_version: TlsVersion::Tls13,
            mutual_tls: false,
            hybrid_mode: true,
        }
    }

    /// Validate that the configuration is self-consistent.
    pub fn validate(&self) -> Result<(), TlsConfigError> {
        if self.preferred_algorithms.is_empty() {
            return Err(TlsConfigError::NoAlgorithms);
        }

        if self.min_tls_version == TlsVersion::Tls12 && self.hybrid_mode {
            return Err(TlsConfigError::IncompatibleVersion(
                "hybrid PQC key exchange requires TLS 1.3".into(),
            ));
        }

        Ok(())
    }

    /// Return the list of cipher suites implied by this configuration.
    pub fn cipher_suites(&self) -> Vec<PqcCipherSuite> {
        if self.hybrid_mode {
            vec![
                PqcCipherSuite::Aes256GcmX25519MlKem768,
                PqcCipherSuite::Aes128GcmX25519MlKem512,
            ]
        } else {
            vec![PqcCipherSuite::Aes256GcmMlKem1024]
        }
    }
}

/// Errors arising from TLS configuration.
#[derive(Debug, thiserror::Error)]
pub enum TlsConfigError {
    #[error("no preferred algorithms specified")]
    NoAlgorithms,
    #[error("incompatible TLS version: {0}")]
    IncompatibleVersion(String),
    #[error("certificate error: {0}")]
    Certificate(String),
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn default_config_is_valid() {
        let cfg = TlsConfig::default();
        assert!(cfg.validate().is_ok());
        assert!(cfg.hybrid_mode);
        assert_eq!(cfg.min_tls_version, TlsVersion::Tls13);
    }

    #[test]
    fn tls12_hybrid_is_invalid() {
        let cfg = TlsConfig {
            min_tls_version: TlsVersion::Tls12,
            hybrid_mode: true,
            ..TlsConfig::default()
        };
        assert!(cfg.validate().is_err());
    }

    #[test]
    fn empty_algorithms_is_invalid() {
        let cfg = TlsConfig {
            preferred_algorithms: vec![],
            ..TlsConfig::default()
        };
        assert!(cfg.validate().is_err());
    }

    #[test]
    fn hybrid_cipher_suites() {
        let cfg = TlsConfig::default();
        let suites = cfg.cipher_suites();
        assert_eq!(suites[0], PqcCipherSuite::Aes256GcmX25519MlKem768);
    }
}
