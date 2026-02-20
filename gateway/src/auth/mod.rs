use axum::{
    body::Body,
    middleware::Next,
    response::{IntoResponse, Response},
};
use http::{Request, StatusCode};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiKey {
    pub id: String,
    pub name: String,
    pub scopes: Vec<String>,
}

#[derive(Debug, Clone)]
pub struct AuthConfig {
    pub require_auth: bool,
    pub api_keys: Vec<ApiKey>,
    pub bypass_paths: Vec<String>,
}

impl Default for AuthConfig {
    fn default() -> Self {
        Self {
            require_auth: false,
            api_keys: Vec::new(),
            bypass_paths: vec!["/health".into(), "/gateway/stats".into()],
        }
    }
}

pub async fn auth_middleware(
    req: Request<Body>,
    next: Next,
) -> Response {
    let config = AuthConfig::default();

    if !config.require_auth {
        return next.run(req).await;
    }

    let path = req.uri().path().to_string();
    if config.bypass_paths.iter().any(|p| path.starts_with(p)) {
        return next.run(req).await;
    }

    let api_key = req
        .headers()
        .get("x-api-key")
        .and_then(|v| v.to_str().ok());

    match api_key {
        Some(key) => {
            if config.api_keys.iter().any(|k| k.id == key) {
                next.run(req).await
            } else {
                (StatusCode::FORBIDDEN, "invalid API key").into_response()
            }
        }
        None => (StatusCode::UNAUTHORIZED, "API key required").into_response(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_auth_config() {
        let config = AuthConfig::default();
        assert!(!config.require_auth);
        assert!(config.bypass_paths.contains(&"/health".to_string()));
    }
}
