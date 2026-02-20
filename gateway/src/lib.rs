pub mod auth;
pub mod middleware;
pub mod proxy;
pub mod tls;

use axum::{routing::get, Router};
use std::net::SocketAddr;

pub struct GatewayConfig {
    pub listen_addr: SocketAddr,
    pub tls_policy: TlsPolicy,
    pub max_connections: usize,
    pub upstream_timeout_secs: u64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum TlsPolicy {
    PqcOnly,
    PqcPreferred,
    Hybrid,
    ClassicalAllowed,
}

impl Default for GatewayConfig {
    fn default() -> Self {
        Self {
            listen_addr: SocketAddr::from(([0, 0, 0, 0], 8443)),
            tls_policy: TlsPolicy::PqcPreferred,
            max_connections: 10_000,
            upstream_timeout_secs: 30,
        }
    }
}

pub fn build_router(config: &GatewayConfig) -> Router {
    Router::new()
        .route("/health", get(health_check))
        .route(
            "/gateway/stats",
            get({
                let policy = config.tls_policy;
                move || stats(policy)
            }),
        )
        .layer(axum::middleware::from_fn_with_state(
            config.tls_policy,
            middleware::pqc_enforcement_middleware,
        ))
        .with_state(config.tls_policy)
}

async fn health_check() -> axum::Json<serde_json::Value> {
    axum::Json(serde_json::json!({
        "status": "ok",
        "service": "qsgw-gateway"
    }))
}

async fn stats(policy: TlsPolicy) -> axum::Json<serde_json::Value> {
    axum::Json(serde_json::json!({
        "tls_policy": format!("{:?}", policy),
        "active_connections": 0,
        "pqc_sessions": 0,
        "classical_sessions": 0,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::body::Body;
    use http::Request;
    use tower::ServiceExt;

    #[tokio::test]
    async fn test_health_check() {
        let config = GatewayConfig::default();
        let app = build_router(&config);

        let response = app
            .oneshot(Request::builder().uri("/health").body(Body::empty()).unwrap())
            .await
            .unwrap();

        assert_eq!(response.status(), 200);
    }

    #[tokio::test]
    async fn test_stats_endpoint() {
        let config = GatewayConfig::default();
        let app = build_router(&config);

        let response = app
            .oneshot(
                Request::builder()
                    .uri("/gateway/stats")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(response.status(), 200);
    }
}
