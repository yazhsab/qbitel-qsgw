use axum::{
    body::Body,
    extract::State,
    middleware::Next,
    response::{IntoResponse, Response},
};
use http::{Request, StatusCode};
use std::time::Instant;
use tracing::info;

use crate::TlsPolicy;

pub async fn pqc_enforcement_middleware(
    State(policy): State<TlsPolicy>,
    req: Request<Body>,
    next: Next,
) -> Response {
    let start = Instant::now();
    let method = req.method().clone();
    let path = req.uri().path().to_string();

    // Check for PQC cipher suite header (set by TLS termination layer)
    let cipher_suite = req
        .headers()
        .get("x-tls-cipher-suite")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("unknown");

    let is_pqc = crate::tls::classify_cipher_suite(cipher_suite);

    if policy == TlsPolicy::PqcOnly && !is_pqc && path != "/health" {
        return (
            StatusCode::FORBIDDEN,
            "PQC-only policy: classical cipher suites not allowed",
        )
            .into_response();
    }

    let response = next.run(req).await;

    let duration = start.elapsed();
    info!(
        method = %method,
        path = %path,
        status = %response.status().as_u16(),
        duration_ms = %duration.as_millis(),
        pqc = is_pqc,
        "request completed"
    );

    response
}

pub async fn rate_limit_middleware(
    req: Request<Body>,
    next: Next,
) -> Response {
    next.run(req).await
}

#[cfg(test)]
mod tests {
    use crate::tls::classify_cipher_suite;

    #[test]
    fn test_pqc_classification_in_middleware() {
        assert!(classify_cipher_suite("TLS_ML-KEM-768_AES_256_GCM"));
        assert!(!classify_cipher_suite("TLS_ECDHE_RSA_AES_256_GCM"));
    }
}
