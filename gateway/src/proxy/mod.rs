use axum::body::Body;
use http::{Request, Response, Uri};
use hyper_util::client::legacy::Client;
use hyper_util::rt::TokioExecutor;
use serde::{Deserialize, Serialize};
use std::time::Duration;
use thiserror::Error;
use tracing::{error, info};

#[derive(Debug, Error)]
pub enum ProxyError {
    #[error("upstream connection failed: {0}")]
    ConnectionFailed(String),
    #[error("upstream timeout")]
    Timeout,
    #[error("no healthy upstream available")]
    NoHealthyUpstream,
    #[error("request error: {0}")]
    RequestError(String),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Upstream {
    pub name: String,
    pub host: String,
    pub port: u16,
    pub is_healthy: bool,
    pub tls_verify: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Route {
    pub path_prefix: String,
    pub upstream: Upstream,
    pub strip_prefix: bool,
    pub priority: i32,
}

pub struct ProxyService {
    routes: Vec<Route>,
    timeout: Duration,
}

impl ProxyService {
    pub fn new(routes: Vec<Route>, timeout_secs: u64) -> Self {
        Self {
            routes,
            timeout: Duration::from_secs(timeout_secs),
        }
    }

    pub fn find_route(&self, path: &str) -> Option<&Route> {
        self.routes
            .iter()
            .filter(|r| path.starts_with(&r.path_prefix) && r.upstream.is_healthy)
            .max_by_key(|r| r.priority)
    }

    pub async fn forward(
        &self,
        route: &Route,
        mut req: Request<Body>,
    ) -> Result<Response<Body>, ProxyError> {
        let upstream_uri = self.build_upstream_uri(route, req.uri())?;
        *req.uri_mut() = upstream_uri;

        // Remove hop-by-hop headers
        let headers = req.headers_mut();
        headers.remove("host");
        headers.remove("connection");

        // Add forwarding headers
        headers.insert(
            "X-Forwarded-Proto",
            "https".parse().unwrap(),
        );

        info!(
            upstream = %route.upstream.name,
            path = %req.uri(),
            "forwarding request"
        );

        let client = Client::builder(TokioExecutor::new()).build_http::<Body>();

        let response = tokio::time::timeout(self.timeout, client.request(req))
            .await
            .map_err(|_| ProxyError::Timeout)?
            .map_err(|e| {
                error!(error = %e, "upstream request failed");
                ProxyError::ConnectionFailed(e.to_string())
            })?;

        // Map the hyper Incoming body to axum Body
        let (parts, incoming) = response.into_parts();
        let body = Body::new(incoming);
        Ok(Response::from_parts(parts, body))
    }

    fn build_upstream_uri(&self, route: &Route, original: &Uri) -> Result<Uri, ProxyError> {
        let path = if route.strip_prefix {
            original
                .path()
                .strip_prefix(&route.path_prefix)
                .unwrap_or(original.path())
        } else {
            original.path()
        };

        let uri_string = format!(
            "http://{}:{}{}",
            route.upstream.host, route.upstream.port, path
        );

        uri_string
            .parse::<Uri>()
            .map_err(|e| ProxyError::RequestError(e.to_string()))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn test_upstream() -> Upstream {
        Upstream {
            name: "test-svc".into(),
            host: "127.0.0.1".into(),
            port: 8080,
            is_healthy: true,
            tls_verify: false,
        }
    }

    #[test]
    fn test_find_route() {
        let routes = vec![
            Route {
                path_prefix: "/api".into(),
                upstream: test_upstream(),
                strip_prefix: false,
                priority: 100,
            },
            Route {
                path_prefix: "/api/v2".into(),
                upstream: test_upstream(),
                strip_prefix: true,
                priority: 200,
            },
        ];

        let svc = ProxyService::new(routes, 30);

        let route = svc.find_route("/api/v2/users").unwrap();
        assert_eq!(route.path_prefix, "/api/v2");

        let route = svc.find_route("/api/v1/keys").unwrap();
        assert_eq!(route.path_prefix, "/api");

        assert!(svc.find_route("/other").is_none());
    }
}
