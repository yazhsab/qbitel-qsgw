# Gateway Configuration Guide

This guide covers the configuration options for the QSGW gateway engine, including TLS policies, routing, upstream management, rate limiting, and performance tuning.

## Table of Contents

- [TLS Policy Configuration](#tls-policy-configuration)
- [Cipher Suite Selection](#cipher-suite-selection)
- [Route Configuration](#route-configuration)
- [Upstream Configuration](#upstream-configuration)
- [Rate Limiting](#rate-limiting)
- [Middleware Pipeline](#middleware-pipeline)
- [Certificate Management](#certificate-management)
- [Performance Tuning](#performance-tuning)

---

## TLS Policy Configuration

QSGW supports four TLS policies that control which cryptographic algorithms the gateway negotiates with clients. The policy is set per gateway instance and determines the level of post-quantum cryptographic enforcement.

### PQC_ONLY

**Maximum security. Requires PQC-capable clients.**

The gateway exclusively negotiates post-quantum cipher suites. Clients that do not support PQC algorithms will have their connections rejected at the TLS handshake level.

| Property          | Value                                     |
|-------------------|-------------------------------------------|
| Key Exchange      | ML-KEM-768, ML-KEM-1024                   |
| Signatures        | ML-DSA-65, ML-DSA-87, SLH-DSA            |
| Minimum TLS       | 1.3                                       |
| Classical Fallback| None                                      |
| Use Case          | High-security environments, PQC mandates  |

```json
{
  "tls_policy": "PQC_ONLY"
}
```

**Considerations:** All connecting clients must support NIST post-quantum standards. This policy provides the strongest quantum resistance but may reject legacy clients.

### PQC_PREFERRED

**Recommended default. PQC with hybrid fallback.**

The gateway prefers post-quantum cipher suites but falls back to hybrid key exchange (classical + PQC) when clients do not support pure PQC. This provides quantum safety while maintaining broad compatibility.

| Property          | Value                                                      |
|-------------------|------------------------------------------------------------|
| Key Exchange      | ML-KEM-768 (preferred), X25519+ML-KEM-768 (fallback)      |
| Signatures        | ML-DSA-65 (preferred), Ed25519+ML-DSA-65 (fallback)       |
| Minimum TLS       | 1.3                                                       |
| Classical Fallback| Hybrid only                                               |
| Use Case          | Production environments, general use                       |

```json
{
  "tls_policy": "PQC_PREFERRED"
}
```

**Considerations:** This is the recommended policy for most deployments. It maximizes quantum safety while ensuring that clients with hybrid support can still connect.

### HYBRID

**Transitional. X25519 + ML-KEM-768 hybrid key exchange.**

The gateway negotiates hybrid key exchange that combines classical elliptic curve cryptography (X25519) with post-quantum lattice-based key encapsulation (ML-KEM-768). This provides defense-in-depth: even if one algorithm is broken, the other provides protection.

| Property          | Value                                          |
|-------------------|-------------------------------------------------|
| Key Exchange      | X25519+ML-KEM-768                              |
| Signatures        | Ed25519+ML-DSA-65                              |
| Minimum TLS       | 1.3                                            |
| Classical Fallback| Embedded in hybrid                             |
| Use Case          | Transitional deployments, compliance            |

```json
{
  "tls_policy": "HYBRID"
}
```

**Considerations:** Suitable for organizations in the early stages of PQC migration that require both classical and quantum-safe protection simultaneously.

### CLASSICAL_ALLOWED

**Legacy compatibility. All TLS 1.2+ cipher suites permitted.**

The gateway allows classical TLS cipher suites alongside PQC suites. PQC suites are still preferred in the negotiation order, but clients may negotiate purely classical connections.

| Property          | Value                                               |
|-------------------|-----------------------------------------------------|
| Key Exchange      | ML-KEM-768, X25519+ML-KEM-768, X25519, ECDHE       |
| Signatures        | ML-DSA-65, Ed25519+ML-DSA-65, Ed25519, ECDSA       |
| Minimum TLS       | 1.2                                                 |
| Classical Fallback| Full classical support                              |
| Use Case          | Legacy environments, migration periods               |

```json
{
  "tls_policy": "CLASSICAL_ALLOWED"
}
```

**Considerations:** This policy maximizes compatibility at the cost of reduced quantum protection. Classical-only connections will trigger `QUANTUM_DOWNGRADE` threat events in the AI threat detection system. Use this policy only during migration periods.

---

## Cipher Suite Selection

The gateway supports the following post-quantum cipher suites based on NIST FIPS standards:

### Key Encapsulation (FIPS 203 -- ML-KEM)

| Algorithm    | Security Level | Key Size  | Ciphertext Size | Use Case              |
|------------- |----------------|-----------|------------------|-----------------------|
| ML-KEM-768  | 3 (AES-192)    | 1,184 B   | 1,088 B          | Recommended default   |
| ML-KEM-1024 | 5 (AES-256)    | 1,568 B   | 1,568 B          | Maximum security      |

### Digital Signatures (FIPS 204 -- ML-DSA)

| Algorithm   | Security Level | Public Key | Signature Size | Use Case              |
|-------------|----------------|------------|----------------|-----------------------|
| ML-DSA-65   | 3 (AES-192)    | 1,952 B    | 3,309 B        | Recommended default   |
| ML-DSA-87   | 5 (AES-256)    | 2,592 B    | 4,627 B        | Maximum security      |

### Stateless Hash-Based Signatures (FIPS 205 -- SLH-DSA)

SLH-DSA provides a conservative, hash-based alternative to lattice-based signatures. It is supported for environments that require signature diversity or distrust lattice assumptions.

### Hybrid Combinations

| Hybrid Key Exchange       | Components               |
|---------------------------|--------------------------|
| X25519+ML-KEM-768        | ECDH + lattice KEM       |

| Hybrid Signatures         | Components               |
|---------------------------|--------------------------|
| Ed25519+ML-DSA-65        | EdDSA + lattice signature|

---

## Route Configuration

Routes define how incoming requests are matched and forwarded to upstream services. Each route belongs to a specific gateway instance.

### Path-Prefix Matching

Routes are matched by longest path prefix. When multiple routes match a request, the route with the highest `priority` value takes precedence.

```
Request: GET /api/users/123/profile

Route A: path_prefix="/api"         priority=0   -> matches
Route B: path_prefix="/api/users"   priority=10  -> matches (longer prefix + higher priority)
Route C: path_prefix="/dashboard"   priority=20  -> no match

Result: Route B is selected
```

### HTTP Method Filtering

Restrict a route to specific HTTP methods. If `methods` is omitted or empty, all methods are allowed.

```json
{
  "path_prefix": "/api/users",
  "methods": ["GET", "POST", "PUT", "DELETE"]
}
```

Requests with non-matching methods receive a `405 Method Not Allowed` response.

### Strip Prefix

When `strip_prefix` is enabled, the matched path prefix is removed before forwarding to the upstream.

```
Route:   path_prefix="/api/v1", strip_prefix=true
Request: GET /api/v1/users/123
Upstream receives: GET /users/123
```

When `strip_prefix` is disabled (the default), the upstream receives the full original path.

### Per-Route Rate Limiting

Each route can define its own requests-per-second limit:

```json
{
  "path_prefix": "/api/search",
  "rate_limit_rps": 50
}
```

A value of `0` (the default) means the route inherits the global rate limit configuration. Per-route limits are enforced independently of per-IP rate limits.

---

## Upstream Configuration

Upstreams represent backend services that the gateway proxies requests to.

### Health Checking

The gateway periodically probes upstream health endpoints to determine availability.

| Parameter           | Default    | Description                              |
|---------------------|------------|------------------------------------------|
| `health_check_path` | `/health`  | HTTP path to probe                       |
| Health interval      | 10s        | Time between health check probes         |
| Healthy threshold    | 2          | Consecutive successes to mark healthy    |
| Unhealthy threshold  | 3          | Consecutive failures to mark unhealthy   |

Health check probes expect an HTTP `200` response. Any other status code (or connection failure) counts as a failure.

### Protocol Selection

Upstreams support multiple protocols for connecting to backend services:

| Protocol | Description                                    |
|----------|------------------------------------------------|
| `HTTP`   | Plain HTTP (suitable for internal services)    |
| `HTTPS`  | TLS-encrypted HTTP                             |
| `GRPC`   | gRPC over HTTP/2                               |
| `TCP`    | Raw TCP proxying                               |
| `TLS`    | Raw TLS proxying                               |

```json
{
  "name": "grpc-service",
  "host": "10.0.2.10",
  "port": 50051,
  "protocol": "GRPC"
}
```

### TLS Verification

When connecting to HTTPS or TLS upstreams, the gateway verifies the upstream's TLS certificate by default.

```json
{
  "tls_verify": true
}
```

Set `tls_verify` to `false` for internal services using self-signed certificates. This is not recommended for production external upstreams.

---

## Rate Limiting

QSGW implements a sliding-window rate limiter to protect both the gateway and upstream services from excessive traffic.

### Per-IP Rate Limiting

Each client IP address is tracked independently with a sliding window counter.

| Parameter               | Default | Description                              |
|-------------------------|---------|------------------------------------------|
| `QSGW_RATE_LIMIT_RPS`  | 100     | Requests per second per IP               |
| `QSGW_RATE_LIMIT_BURST`| 200     | Maximum burst size                       |
| Window size             | 1s      | Sliding window duration                  |
| Cleanup interval        | 60s     | Stale entry cleanup frequency            |

### Per-Route Rate Limiting

Routes can override the global rate limit with a custom value via the `rate_limit_rps` field. Per-route limits apply in addition to per-IP limits. A request must satisfy both limits to proceed.

### Rate Limit Headers

When rate limiting is active, the gateway includes the following response headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 42
X-RateLimit-Reset: 1708425600
```

When the limit is exceeded, the gateway returns `429 Too Many Requests` with a `Retry-After` header.

---

## Middleware Pipeline

Every request processed by the gateway passes through the following middleware stages in order:

```
Client Request
  |
  v
[1] TLS Termination      -- PQC/hybrid/classical TLS handshake
  |
  v
[2] Authentication        -- JWT or API key validation
  |
  v
[3] Rate Limiting         -- Per-IP and per-route rate checks
  |
  v
[4] PQC Enforcement       -- TLS policy compliance check; threat generation
  |
  v
[5] Route Matching        -- Path-prefix + priority + method matching
  |
  v
[6] Proxy                 -- Forward to upstream, connection pooling
  |
  v
[7] Response Headers      -- Security headers (HSTS, CSP, X-Frame-Options)
  |
  v
Client Response
```

### Stage Details

1. **TLS Termination:** The Rust gateway engine (rustls) terminates the TLS connection according to the gateway's TLS policy. Session details are recorded in the `tls_sessions` table.

2. **Authentication:** JWT tokens are validated for expiry and signature (HMAC-SHA256). API keys are verified using constant-time comparison. Unauthenticated requests to protected endpoints receive `401 Unauthorized`.

3. **Rate Limiting:** The sliding-window rate limiter checks both per-IP and per-route limits. Requests that exceed limits receive `429 Too Many Requests`.

4. **PQC Enforcement:** The gateway checks whether the negotiated cipher suite complies with the configured TLS policy. Non-compliant connections generate threat events (e.g., `QUANTUM_DOWNGRADE`).

5. **Route Matching:** The request path is matched against configured routes by longest prefix, then by priority. Unmatched requests receive `404 Not Found`.

6. **Proxy:** The request is forwarded to the selected upstream service. Connection pooling reduces overhead for repeated requests to the same upstream.

7. **Response Headers:** Security headers are injected into every response:
   - `Strict-Transport-Security: max-age=63072000; includeSubDomains`
   - `Content-Security-Policy: default-src 'self'`
   - `X-Frame-Options: DENY`
   - `X-Content-Type-Options: nosniff`
   - `X-XSS-Protection: 0`
   - `Referrer-Policy: strict-origin-when-cross-origin`

---

## Certificate Management

### TLS Certificate and Key Paths

The gateway expects TLS certificates and private keys at the following default locations:

| Environment Variable    | Default                       | Description           |
|-------------------------|-------------------------------|-----------------------|
| `QSGW_TLS_CERT_PATH`   | `/etc/qsgw/certs/server.crt` | PEM-encoded certificate chain |
| `QSGW_TLS_KEY_PATH`    | `/etc/qsgw/certs/server.key` | PEM-encoded private key       |

The certificate chain should include the server certificate followed by any intermediate certificates.

### Certificate Rotation

QSGW supports graceful certificate rotation without downtime:

1. Place the new certificate and key files at the configured paths.
2. Send a `SIGHUP` signal to the gateway process, or call the control plane reload endpoint.
3. The gateway reloads the certificate for all new connections. Existing connections continue using the previous certificate until they close.

### PQC Certificate Considerations

For full post-quantum TLS, use certificates signed with ML-DSA or SLH-DSA algorithms. Hybrid certificates (Ed25519 + ML-DSA) are supported for transitional deployments.

---

## Performance Tuning

### Connection Limits

| Environment Variable         | Default | Description                              |
|------------------------------|---------|------------------------------------------|
| `QSGW_MAX_CONNECTIONS`       | 10000   | Maximum concurrent client connections    |
| `QSGW_UPSTREAM_POOL_SIZE`    | 100     | Connections per upstream in the pool     |
| `QSGW_UPSTREAM_TIMEOUT_SECS` | 30      | Upstream request timeout in seconds      |

### Worker Threads

The Rust gateway uses Tokio's multi-threaded runtime. By default, it spawns one worker thread per CPU core.

| Environment Variable         | Default     | Description                          |
|------------------------------|-------------|--------------------------------------|
| `QSGW_WORKER_THREADS`       | CPU cores   | Number of Tokio worker threads       |
| `QSGW_BLOCKING_THREADS`     | 512         | Maximum blocking thread pool size    |

### Connection Pooling

Upstream connection pooling reduces the overhead of establishing new connections for each proxied request.

| Parameter              | Default | Description                              |
|------------------------|---------|------------------------------------------|
| Pool size per upstream | 100     | Maximum idle connections per upstream    |
| Idle timeout           | 90s     | Time before idle connections are closed  |
| Connection lifetime    | 300s    | Maximum connection lifetime              |

### Recommended Production Settings

For a high-traffic production deployment handling 10,000+ requests per second:

```bash
# Gateway performance
export QSGW_MAX_CONNECTIONS=50000
export QSGW_WORKER_THREADS=16
export QSGW_UPSTREAM_POOL_SIZE=200
export QSGW_UPSTREAM_TIMEOUT_SECS=15

# Rate limiting
export QSGW_RATE_LIMIT_RPS=500
export QSGW_RATE_LIMIT_BURST=1000
```

### Memory Considerations

PQC cipher suites have larger key and ciphertext sizes compared to classical algorithms. This increases per-connection memory usage:

| TLS Policy         | Approximate Memory per Connection |
|--------------------|-----------------------------------|
| CLASSICAL_ALLOWED  | ~2 KB                             |
| HYBRID             | ~4 KB                             |
| PQC_PREFERRED      | ~5 KB                             |
| PQC_ONLY           | ~6 KB                             |

Plan memory allocation accordingly when setting `QSGW_MAX_CONNECTIONS`. For example, 50,000 connections with `PQC_PREFERRED` requires approximately 250 MB for TLS state alone.
