use quantun_tls::config::{TlsConfig, TlsVersion};
use quantun_types::algorithm::{MlKemVariant, MlDsaVariant};
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tracing::info;

use crate::TlsPolicy;

#[derive(Debug, Error)]
pub enum TlsError {
    #[error("no PQC cipher suites available")]
    NoPqcCipherSuites,
    #[error("TLS policy violation: {0}")]
    PolicyViolation(String),
    #[error("configuration error: {0}")]
    ConfigError(String),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HandshakeInfo {
    pub cipher_suite: String,
    pub tls_version: String,
    pub kem_algorithm: Option<String>,
    pub sig_algorithm: Option<String>,
    pub is_pqc: bool,
    pub handshake_duration_ms: u64,
}

pub fn build_tls_config(policy: TlsPolicy) -> Result<TlsConfig, TlsError> {
    let mut config = TlsConfig::development();
    config.min_tls_version = TlsVersion::Tls13;

    match policy {
        TlsPolicy::PqcOnly => {
            config.preferred_algorithms = vec![
                quantun_types::Algorithm::MlKem(MlKemVariant::MlKem768),
                quantun_types::Algorithm::MlDsa(MlDsaVariant::MlDsa65),
            ];
            config.hybrid_mode = false;
            info!("TLS configured: PQC-only mode");
        }
        TlsPolicy::PqcPreferred => {
            config.preferred_algorithms = vec![
                quantun_types::Algorithm::MlKem(MlKemVariant::MlKem768),
                quantun_types::Algorithm::MlKem(MlKemVariant::MlKem1024),
                quantun_types::Algorithm::MlDsa(MlDsaVariant::MlDsa65),
            ];
            config.hybrid_mode = true;
            info!("TLS configured: PQC-preferred mode (hybrid enabled)");
        }
        TlsPolicy::Hybrid => {
            config.preferred_algorithms = vec![
                quantun_types::Algorithm::MlKem(MlKemVariant::MlKem768),
            ];
            config.hybrid_mode = true;
            info!("TLS configured: Hybrid mode");
        }
        TlsPolicy::ClassicalAllowed => {
            config.preferred_algorithms = vec![];
            config.hybrid_mode = false;
            info!("TLS configured: Classical allowed mode");
        }
    }

    config.validate().map_err(|e| TlsError::ConfigError(e.to_string()))?;
    Ok(config)
}

pub fn classify_cipher_suite(cipher_suite: &str) -> bool {
    let pqc_indicators = ["ML-KEM", "ML-DSA", "SLH-DSA", "KYBER", "DILITHIUM"];
    pqc_indicators.iter().any(|p| cipher_suite.contains(p))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_build_tls_config_pqc_preferred() {
        let config = build_tls_config(TlsPolicy::PqcPreferred).unwrap();
        assert_eq!(config.min_tls_version, TlsVersion::Tls13);
        assert!(config.hybrid_mode);
        assert!(!config.preferred_algorithms.is_empty());
    }

    #[test]
    fn test_classify_cipher_suite() {
        assert!(classify_cipher_suite("TLS_ML-KEM-768_AES_256_GCM_SHA384"));
        assert!(!classify_cipher_suite("TLS_AES_256_GCM_SHA384"));
    }
}
