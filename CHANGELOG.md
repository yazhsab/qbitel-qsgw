# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-20

### Added
- Initial release of QSGW as a standalone open-source product
- Rust gateway engine with Axum for async reverse proxying and PQC TLS termination
- Post-quantum TLS support: ML-KEM (FIPS 203), ML-DSA (FIPS 204), SLH-DSA (FIPS 205)
- Hybrid key exchange (X25519 + ML-KEM-768) and hybrid signatures (Ed25519 + ML-DSA-65)
- Four TLS policies: PQC_ONLY, PQC_PREFERRED, HYBRID, CLASSICAL_ALLOWED
- Go control-plane REST API for managing gateways, routes, upstreams, and threat events
- Python AI engine with anomaly detection (quantum downgrade, traffic spikes) and bot detection
- React 19 admin dashboard with gateway, upstream, route, and threat monitoring
- JWT (HMAC-SHA256) and API key authentication with constant-time comparison
- Per-IP rate limiting with sliding window and automatic cleanup
- Security headers middleware (HSTS, CSP, X-Frame-Options)
- TLS session tracking and cipher suite analysis
- PostgreSQL schema with gateways, upstreams, routes, tls_sessions, threat_events, and audit_log
- etcd integration for distributed configuration
- Docker Compose configuration for local development and deployment
- GitHub Actions CI pipeline (Rust, Go, Python, TypeScript)
- Comprehensive API, gateway configuration, and threat detection documentation
- Apache 2.0 license
