<p align="center">
  <strong>QSGW</strong> &mdash; Quantum-Safe Gateway
</p>

<p align="center">
  <a href="https://github.com/quantun-opensource/qsgw/actions"><img src="https://github.com/quantun-opensource/qsgw/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://www.rust-lang.org"><img src="https://img.shields.io/badge/rust-1.75%2B-orange.svg" alt="Rust"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/go-1.23-00ADD8.svg" alt="Go"></a>
  <a href="https://python.org"><img src="https://img.shields.io/badge/python-3.11%2B-3776AB.svg" alt="Python"></a>
  <a href="https://react.dev"><img src="https://img.shields.io/badge/react-19-61DAFB.svg" alt="React"></a>
</p>

---

**Drop-in quantum-safe replacement for Kong, Apigee, and traditional API gateways.**

QSGW is a post-quantum cryptographic API gateway that protects your services against both classical and quantum-computing threats. It terminates TLS with FIPS 203/204/205 compliant post-quantum algorithms, reverse-proxies traffic to your upstreams, and uses AI-powered threat detection to identify quantum downgrade attacks, anomalous traffic, and bot activity -- all in a single deployment.

---

## Key Features

- **Post-Quantum TLS Termination** -- FIPS 203 (ML-KEM), FIPS 204 (ML-DSA), FIPS 205 (SLH-DSA) with hybrid X25519 + ML-KEM-768 key exchange. Four TLS policies to match your migration timeline.
- **High-Performance Reverse Proxy** -- Built in Rust with Axum and Tokio. Async I/O, zero-copy forwarding, hop-by-hop header stripping, and configurable upstream timeouts.
- **AI-Powered Threat Detection** -- Python FastAPI service with anomaly detection (quantum downgrade attacks, traffic spikes, abnormal handshakes) and bot detection (user-agent fingerprinting, rate analysis, path diversity scoring).
- **Control Plane REST API** -- Go service managing gateways, routes, upstreams, and threat events. JWT and API key authentication with constant-time comparison.
- **Admin Dashboard** -- React 19 + TypeScript single-page application for monitoring gateways, upstreams, routes, and threat events.
- **Flexible Authentication** -- JWT (HS256) and API key authentication with configurable bypass paths and role-based access control.
- **Per-IP Rate Limiting** -- In-memory sliding window rate limiter with configurable thresholds, automatic cleanup, and memory-bounded key tracking.
- **Security Headers** -- HSTS, CSP, X-Frame-Options, and other hardened response headers out of the box.
- **Configurable TLS Policies** -- PQC_ONLY, PQC_PREFERRED, HYBRID, and CLASSICAL_ALLOWED to support gradual migration.

## Architecture

```
                         Internet
                            |
                            v
                 +---------------------+
                 |   QSGW Gateway      |  Rust (Axum) -- port 8443
                 |   PQC TLS Termination|
                 |   Reverse Proxy     |
                 +----------+----------+
                            |
           +----------------+----------------+
           |                |                |
           v                v                v
  +-----------------+ +-----------+ +------------------+
  | Control Plane   | | AI Engine | |  Upstream        |
  | Go (Chi)        | | Python    | |  Services        |
  | port 8085       | | (FastAPI) | |  (your backends) |
  +--------+--------+ | port 8086 | +------------------+
           |           +-----------+
           v
  +-----------------+    +---------+
  | PostgreSQL      |    |  etcd   |
  | port 5432       |    |  2379   |
  +-----------------+    +---------+
           |
           v
  +-----------------+
  |  Admin UI       |  React 19 -- port 3003
  +-----------------+
```

**Traffic flow**: Clients connect to the Gateway over PQC TLS (port 8443). The Gateway terminates TLS, enforces cipher suite policy, applies rate limiting and authentication, then reverse-proxies requests to upstream services. The Control Plane manages configuration. The AI Engine analyzes traffic for threats.

## TLS Policy Options

| Policy | Description | Use Case |
|--------|-------------|----------|
| `PQC_ONLY` | Only post-quantum cipher suites accepted. Classical connections rejected. | Maximum quantum safety. Requires PQC-capable clients. |
| `PQC_PREFERRED` | PQC cipher suites preferred, hybrid fallback enabled. | Recommended default. Best balance of safety and compatibility. |
| `HYBRID` | X25519 + ML-KEM-768 hybrid key exchange. | Transitional deployments with mixed clients. |
| `CLASSICAL_ALLOWED` | PQC available but classical cipher suites permitted. | Legacy compatibility during migration. |

## Supported Cryptographic Algorithms

| Standard | Algorithm | Variants | Purpose |
|----------|-----------|----------|---------|
| FIPS 203 | ML-KEM | ML-KEM-512, ML-KEM-768, ML-KEM-1024 | Key Encapsulation Mechanism |
| FIPS 204 | ML-DSA | ML-DSA-44, ML-DSA-65, ML-DSA-87 | Digital Signatures |
| FIPS 205 | SLH-DSA | SHA2-128s/f, SHA2-192s/f, SHA2-256s/f | Stateless Hash-Based Signatures |
| Hybrid | X25519 + ML-KEM-768 | -- | Hybrid Key Exchange |
| Hybrid | Ed25519 + ML-DSA-65 | -- | Hybrid Signatures |

## Quick Start

### Docker (recommended)

```bash
# Clone the repository
git clone https://github.com/quantun-opensource/qsgw.git
cd qsgw

# Copy the example environment configuration
cp .env.example .env

# Start all services (Gateway, Control Plane, AI Engine, Admin UI, PostgreSQL, etcd)
make docker-all

# Verify services are running
curl http://localhost:8085/health    # Control Plane
curl http://localhost:8086/health    # AI Engine
```

| Service | URL | Description |
|---------|-----|-------------|
| Gateway | `https://localhost:8443` | PQC TLS reverse proxy |
| Control Plane | `http://localhost:8085` | REST API |
| AI Engine | `http://localhost:8086` | Threat analysis |
| Admin UI | `http://localhost:3003` | Dashboard |

### Local Development

```bash
# 1. Install prerequisites
#    Rust >= 1.75, Go >= 1.23, Python >= 3.11, Node.js >= 22, Docker

# 2. Clone and set up dependencies
git clone https://github.com/quantun-opensource/qsgw.git
cd qsgw
make setup

# 3. Start PostgreSQL and etcd
make docker-deps

# 4. Run database migrations
export QSGW_DATABASE_URL="postgres://quantun:quantun_dev@localhost:5432/qsgw?sslmode=disable"
make migrate

# 5. Build all components
make build

# 6. Run all tests
make test
```

## API Overview

### Control Plane (port 8085)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/gateways` | Create gateway |
| `GET` | `/api/v1/gateways` | List gateways |
| `GET` | `/api/v1/gateways/{id}` | Get gateway by ID |
| `POST` | `/api/v1/gateways/{id}/activate` | Activate gateway |
| `POST` | `/api/v1/gateways/{id}/deactivate` | Deactivate gateway |
| `POST` | `/api/v1/upstreams` | Create upstream |
| `GET` | `/api/v1/upstreams` | List upstreams |
| `GET` | `/api/v1/upstreams/{id}` | Get upstream by ID |
| `POST` | `/api/v1/routes` | Create route |
| `GET` | `/api/v1/routes?gateway_id=...` | List routes by gateway |
| `DELETE` | `/api/v1/routes/{id}` | Delete route |
| `GET` | `/api/v1/threats?gateway_id=...` | List threat events |

### AI Engine (port 8086)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/analyze-traffic` | Anomaly detection |
| `POST` | `/api/v1/detect-bot` | Bot detection |

See [docs/API.md](docs/API.md) for complete request/response examples.

## Threat Detection

QSGW identifies the following threat categories:

| Threat Type | Severity | Description |
|-------------|----------|-------------|
| `QUANTUM_DOWNGRADE` | CRITICAL | Attempts to force classical cipher suites, exposing traffic to future quantum attacks |
| `WEAK_CIPHER` | HIGH | Use of deprecated or weak cipher suites (RC4, DES, 3DES, MD5, SHA1) |
| `BOT_ATTACK` | HIGH | Automated traffic attempting data harvesting or endpoint probing |
| `ANOMALOUS_TRAFFIC` | MEDIUM | Unusual traffic patterns, rate spikes, or abnormal handshake durations |
| `CERTIFICATE_ISSUE` | MEDIUM | Invalid, expired, or misconfigured TLS certificates |
| `REPLAY_ATTACK` | HIGH | Detected replay of previously captured TLS sessions |

See [docs/THREAT_DETECTION.md](docs/THREAT_DETECTION.md) for details.

## Configuration

All configuration is via environment variables. See `.env.example` for the full list.

| Variable | Default | Description |
|----------|---------|-------------|
| `QSGW_PORT` | `8085` | Control plane HTTP port |
| `QSGW_DATABASE_URL` | -- | PostgreSQL connection string (required) |
| `QSGW_GATEWAY_PORT` | `8443` | Gateway TLS port |
| `QSGW_AI_ENGINE_PORT` | `8086` | AI engine HTTP port |
| `QSGW_ADMIN_PORT` | `3003` | Admin UI port |
| `QSGW_ETCD_ADDR` | `http://127.0.0.1:2379` | etcd endpoint |
| `QSGW_JWT_SECRET` | -- | JWT signing secret (enables JWT auth) |
| `QSGW_JWT_ISSUER` | `quantun` | Expected JWT issuer claim |
| `QSGW_API_KEYS` | -- | Comma-separated `key:subject:role` entries |
| `QSGW_CORS_ORIGINS` | -- | Allowed CORS origins (comma-separated) |
| `QSGW_MAX_BODY_BYTES` | `1048576` | Maximum request body size (1 MB) |
| `QSGW_LOG_LEVEL` | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |
| `ETCD_ENDPOINTS` | -- | etcd cluster endpoints |

## Project Structure

```
qsgw/
  gateway/              Rust -- PQC-enabled reverse proxy (Axum + Tokio)
    src/
      auth/               API key authentication middleware
      middleware/          PQC enforcement, rate limiting
      proxy/              Reverse proxy with upstream routing
      tls/                TLS policy configuration, cipher suite classification
  crypto/               Rust -- Post-quantum cryptography primitives
    src/
      mlkem.rs            ML-KEM key encapsulation (FIPS 203)
      mldsa.rs            ML-DSA digital signatures (FIPS 204)
      slhdsa.rs           SLH-DSA hash-based signatures (FIPS 205)
      hybrid.rs           Hybrid classical + PQC schemes
      rng.rs              Secure random number generation
  tls/                  Rust -- TLS configuration with PQC cipher suites
  types/                Rust -- Shared type definitions (algorithms, error codes)
  control-plane/        Go -- REST API for gateway management
    cmd/server/           Application entry point
    internal/
      config/             Environment-based configuration
      handler/            HTTP handlers (gateway, route, upstream, threat)
      model/              Domain models and request/response types
      repository/         PostgreSQL data access layer
      service/            Business logic layer
  shared/go/            Go -- Shared libraries
    database/             PostgreSQL connection pool
    middleware/           Auth (JWT + API key), rate limiting, security headers, CORS, pagination
  ai-engine/            Python -- AI-powered threat detection
    src/qsgw_ai/
      anomaly_detector/   Traffic anomaly detection engine
      bot_detector/       Bot and automated traffic detection engine
      api/                FastAPI application
  admin/                TypeScript -- React admin dashboard
    src/
      pages/              Dashboard, Gateways, Upstreams, Threats
      components/         Shared UI components (Card, Button, DataTable)
  db/migrations/        SQL -- PostgreSQL schema migrations
  infra/docker/         Docker Compose configuration
  .github/workflows/    CI/CD pipeline (Rust, Go, Python, TypeScript)
```

## Performance Characteristics

- **Gateway**: Async I/O via Tokio with configurable connection limits (default 10,000 concurrent connections). Zero-copy request forwarding with configurable upstream timeouts (default 30 seconds).
- **Rate Limiting**: In-memory sliding window with O(1) per-request overhead. Background cleanup prevents unbounded memory growth. Capped at 100,000 tracked keys.
- **TLS**: PQC key exchange adds approximately 1-3ms to handshake time compared to classical ECDHE. Hybrid mode (X25519 + ML-KEM-768) provides both classical and quantum security.
- **AI Engine**: Rule-based anomaly and bot detection with sub-millisecond evaluation per request.

## Development

```bash
# Build all components
make build          # Build Rust + Go + Node
make build-rust     # Rust workspace only
make build-go       # Go control-plane only
make build-node     # Admin UI only

# Run tests
make test           # All tests
make test-rust      # Rust unit tests
make test-go        # Go unit tests
make test-python    # Python unit tests

# Lint
make lint           # All linters
make lint-rust      # cargo clippy + cargo fmt --check
make lint-go        # golangci-lint
make lint-python    # ruff

# Database
make migrate        # Run migrations
make migrate-down   # Rollback migrations

# Docker
make docker-deps    # Start PostgreSQL + etcd
make docker-all     # Start all services
make docker-down    # Stop all services

# Clean
make clean          # Remove build artifacts
```

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for the full developer guide.

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/ARCHITECTURE.md) | System design, data flow, and component interactions |
| [API Reference](docs/API.md) | Complete API documentation with request/response examples |
| [Gateway Configuration](docs/GATEWAY_CONFIG.md) | TLS policies, routes, upstreams, and middleware pipeline |
| [Deployment Guide](docs/DEPLOYMENT.md) | Docker, Kubernetes, production setup, and scaling |
| [Development Guide](docs/DEVELOPMENT.md) | Local setup, building, testing, and extending |
| [Threat Detection](docs/THREAT_DETECTION.md) | Anomaly detection, bot detection, and threat types |
| [Contributing](CONTRIBUTING.md) | How to contribute to QSGW |
| [Security Policy](SECURITY.md) | Vulnerability reporting and cryptographic standards |
| [Changelog](CHANGELOG.md) | Release history |

## Contributing

We welcome contributions. Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on submitting pull requests, reporting issues, and development workflow.

## Security

QSGW takes security seriously. If you discover a security vulnerability, please report it responsibly. See [SECURITY.md](SECURITY.md) for our security policy and vulnerability reporting process.

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.

Copyright 2026 Quantun.
