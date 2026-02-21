# QSGW API Reference

Complete API documentation for the Quantum-Safe Gateway control plane and AI engine.

## Table of Contents

- [Authentication](#authentication)
- [Error Format](#error-format)
- [Pagination](#pagination)
- [Control Plane API (Port 8085)](#control-plane-api-port-8085)
  - [Health](#health)
  - [Gateways](#gateways)
  - [Upstreams](#upstreams)
  - [Routes](#routes)
  - [Threats](#threats)
- [AI Engine API (Port 8086)](#ai-engine-api-port-8086)
  - [Health](#ai-engine-health)
  - [Traffic Analysis](#traffic-analysis)
  - [Bot Detection](#bot-detection)

---

## Authentication

All API endpoints (except health checks) require authentication via one of two methods:

### JWT Bearer Token

Include a JWT token signed with HMAC-SHA256 in the `Authorization` header:

```
Authorization: Bearer <jwt-token>
```

**JWT Claims:**

| Claim | Type   | Description              |
|-------|--------|--------------------------|
| `sub` | string | Subject (user or service) |
| `exp` | int    | Expiration timestamp      |
| `iat` | int    | Issued-at timestamp       |
| `iss` | string | Issuer (`qsgw`)           |

**Example -- obtain and use a token:**

```bash
# Use the JWT token in subsequent requests
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:8085/api/v1/gateways
```

### API Key

Include an API key in the `X-API-Key` header:

```
X-API-Key: <api-key>
```

**Example:**

```bash
curl -H "X-API-Key: qsgw_k_abc123def456" \
  http://localhost:8085/api/v1/gateways
```

API keys are validated using constant-time comparison to prevent timing attacks.

---

## Error Format

All error responses follow a consistent JSON structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Field 'name' is required",
    "details": {
      "field": "name",
      "constraint": "required"
    }
  }
}
```

**Standard Error Codes:**

| HTTP Status | Code                  | Description                          |
|-------------|-----------------------|--------------------------------------|
| 400         | `VALIDATION_ERROR`    | Invalid request body or parameters   |
| 401         | `UNAUTHORIZED`        | Missing or invalid credentials       |
| 403         | `FORBIDDEN`           | Insufficient permissions             |
| 404         | `NOT_FOUND`           | Resource does not exist              |
| 409         | `CONFLICT`            | Resource already exists              |
| 429         | `RATE_LIMITED`        | Too many requests                    |
| 500         | `INTERNAL_ERROR`      | Unexpected server error              |

---

## Pagination

List endpoints support cursor-based pagination with the following query parameters:

| Parameter | Type   | Default | Description                        |
|-----------|--------|---------|------------------------------------|
| `limit`   | int    | 20      | Number of items per page (max 100) |
| `offset`  | int    | 0       | Number of items to skip            |

**Paginated Response Envelope:**

```json
{
  "data": [...],
  "pagination": {
    "total": 42,
    "limit": 20,
    "offset": 0
  }
}
```

---

## Control Plane API (Port 8085)

Base URL: `http://localhost:8085`

### Health

#### `GET /health`

Returns the health status of the control plane service. No authentication required.

**Response `200 OK`:**

```json
{
  "status": "healthy",
  "service": "qsgw-control-plane",
  "version": "0.1.0",
  "timestamp": "2026-02-20T10:30:00Z"
}
```

---

### Gateways

#### `POST /api/v1/gateways`

Create a new gateway instance.

**Request Body:**

| Field             | Type   | Required | Default            | Description                                |
|-------------------|--------|----------|--------------------|--------------------------------------------|
| `name`            | string | yes      | --                 | Unique gateway name                        |
| `hostname`        | string | yes      | --                 | Hostname or IP to bind                     |
| `port`            | int    | yes      | --                 | Port to listen on                          |
| `tls_policy`      | string | yes      | --                 | One of: `PQC_ONLY`, `PQC_PREFERRED`, `HYBRID`, `CLASSICAL_ALLOWED` |
| `max_connections` | int    | no       | 10000              | Maximum concurrent connections             |

**Example Request:**

```bash
curl -X POST http://localhost:8085/api/v1/gateways \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-gateway",
    "hostname": "0.0.0.0",
    "port": 8443,
    "tls_policy": "PQC_PREFERRED",
    "max_connections": 50000
  }'
```

**Response `201 Created`:**

```json
{
  "id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
  "name": "production-gateway",
  "hostname": "0.0.0.0",
  "port": 8443,
  "tls_policy": "PQC_PREFERRED",
  "max_connections": 50000,
  "status": "inactive",
  "created_at": "2026-02-20T10:30:00Z",
  "updated_at": "2026-02-20T10:30:00Z"
}
```

---

#### `GET /api/v1/gateways`

List all gateways with pagination.

**Query Parameters:**

| Parameter | Type | Default | Description            |
|-----------|------|---------|------------------------|
| `limit`   | int  | 20      | Items per page         |
| `offset`  | int  | 0       | Items to skip          |

**Example Request:**

```bash
curl http://localhost:8085/api/v1/gateways?limit=10&offset=0 \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "data": [
    {
      "id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
      "name": "production-gateway",
      "hostname": "0.0.0.0",
      "port": 8443,
      "tls_policy": "PQC_PREFERRED",
      "max_connections": 50000,
      "status": "active",
      "created_at": "2026-02-20T10:30:00Z",
      "updated_at": "2026-02-20T10:35:00Z"
    }
  ],
  "pagination": {
    "total": 1,
    "limit": 10,
    "offset": 0
  }
}
```

---

#### `GET /api/v1/gateways/{id}`

Retrieve a single gateway by ID.

**Example Request:**

```bash
curl http://localhost:8085/api/v1/gateways/gw_01HQ3K5V8N2M4P6R7S9T0UVW \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
  "name": "production-gateway",
  "hostname": "0.0.0.0",
  "port": 8443,
  "tls_policy": "PQC_PREFERRED",
  "max_connections": 50000,
  "status": "active",
  "created_at": "2026-02-20T10:30:00Z",
  "updated_at": "2026-02-20T10:35:00Z"
}
```

**Response `404 Not Found`:**

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Gateway not found"
  }
}
```

---

#### `POST /api/v1/gateways/{id}/activate`

Activate a gateway, starting TLS termination and request proxying.

**Example Request:**

```bash
curl -X POST http://localhost:8085/api/v1/gateways/gw_01HQ3K5V8N2M4P6R7S9T0UVW/activate \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
  "status": "active",
  "message": "Gateway activated successfully"
}
```

---

#### `POST /api/v1/gateways/{id}/deactivate`

Deactivate a gateway, gracefully draining existing connections.

**Example Request:**

```bash
curl -X POST http://localhost:8085/api/v1/gateways/gw_01HQ3K5V8N2M4P6R7S9T0UVW/deactivate \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
  "status": "inactive",
  "message": "Gateway deactivated successfully"
}
```

---

### Upstreams

#### `POST /api/v1/upstreams`

Register a new upstream service.

**Request Body:**

| Field               | Type   | Required | Default  | Description                                    |
|---------------------|--------|----------|----------|------------------------------------------------|
| `name`              | string | yes      | --       | Unique upstream name                           |
| `host`              | string | yes      | --       | Upstream host address                          |
| `port`              | int    | yes      | --       | Upstream port                                  |
| `protocol`          | string | no       | `HTTP`   | One of: `HTTP`, `HTTPS`, `GRPC`, `TCP`, `TLS` |
| `tls_verify`        | bool   | no       | `true`   | Verify upstream TLS certificate                |
| `health_check_path` | string | no       | `/health`| Path for health check probes                   |

**Example Request:**

```bash
curl -X POST http://localhost:8085/api/v1/upstreams \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "user-service",
    "host": "10.0.1.50",
    "port": 8080,
    "protocol": "HTTP",
    "tls_verify": false,
    "health_check_path": "/healthz"
  }'
```

**Response `201 Created`:**

```json
{
  "id": "ups_01HQ3M7X9P4Q6R8S0T2U3VWX",
  "name": "user-service",
  "host": "10.0.1.50",
  "port": 8080,
  "protocol": "HTTP",
  "tls_verify": false,
  "health_check_path": "/healthz",
  "health_status": "unknown",
  "created_at": "2026-02-20T10:40:00Z",
  "updated_at": "2026-02-20T10:40:00Z"
}
```

---

#### `GET /api/v1/upstreams`

List all registered upstreams with pagination.

**Example Request:**

```bash
curl http://localhost:8085/api/v1/upstreams?limit=20&offset=0 \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "data": [
    {
      "id": "ups_01HQ3M7X9P4Q6R8S0T2U3VWX",
      "name": "user-service",
      "host": "10.0.1.50",
      "port": 8080,
      "protocol": "HTTP",
      "tls_verify": false,
      "health_check_path": "/healthz",
      "health_status": "healthy",
      "created_at": "2026-02-20T10:40:00Z",
      "updated_at": "2026-02-20T10:40:00Z"
    }
  ],
  "pagination": {
    "total": 1,
    "limit": 20,
    "offset": 0
  }
}
```

---

#### `GET /api/v1/upstreams/{id}`

Retrieve a single upstream by ID.

**Example Request:**

```bash
curl http://localhost:8085/api/v1/upstreams/ups_01HQ3M7X9P4Q6R8S0T2U3VWX \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "id": "ups_01HQ3M7X9P4Q6R8S0T2U3VWX",
  "name": "user-service",
  "host": "10.0.1.50",
  "port": 8080,
  "protocol": "HTTP",
  "tls_verify": false,
  "health_check_path": "/healthz",
  "health_status": "healthy",
  "created_at": "2026-02-20T10:40:00Z",
  "updated_at": "2026-02-20T10:40:00Z"
}
```

---

### Routes

#### `POST /api/v1/routes`

Create a routing rule that maps incoming requests to an upstream service.

**Request Body:**

| Field            | Type     | Required | Default | Description                                |
|------------------|----------|----------|---------|--------------------------------------------|
| `gateway_id`     | string   | yes      | --      | Gateway this route belongs to              |
| `upstream_id`    | string   | yes      | --      | Target upstream service                    |
| `path_prefix`    | string   | yes      | --      | URL path prefix to match (e.g., `/api/v1`) |
| `methods`        | string[] | no       | all     | Allowed HTTP methods                       |
| `strip_prefix`   | bool     | no       | `false` | Strip the matched prefix before proxying   |
| `priority`       | int      | no       | 0       | Higher priority routes match first         |
| `rate_limit_rps` | int      | no       | 0       | Per-route rate limit (0 = unlimited)       |

**Example Request:**

```bash
curl -X POST http://localhost:8085/api/v1/routes \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "gateway_id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
    "upstream_id": "ups_01HQ3M7X9P4Q6R8S0T2U3VWX",
    "path_prefix": "/api/users",
    "methods": ["GET", "POST", "PUT", "DELETE"],
    "strip_prefix": true,
    "priority": 10,
    "rate_limit_rps": 100
  }'
```

**Response `201 Created`:**

```json
{
  "id": "rt_01HQ3N9Y0Q5R7S9T1U3V4WXY",
  "gateway_id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
  "upstream_id": "ups_01HQ3M7X9P4Q6R8S0T2U3VWX",
  "path_prefix": "/api/users",
  "methods": ["GET", "POST", "PUT", "DELETE"],
  "strip_prefix": true,
  "priority": 10,
  "rate_limit_rps": 100,
  "created_at": "2026-02-20T10:45:00Z",
  "updated_at": "2026-02-20T10:45:00Z"
}
```

---

#### `GET /api/v1/routes`

List routes, optionally filtered by gateway.

**Query Parameters:**

| Parameter    | Type   | Required | Description            |
|------------- |--------|----------|------------------------|
| `gateway_id` | string | no       | Filter by gateway ID   |
| `limit`      | int    | no       | Items per page         |
| `offset`     | int    | no       | Items to skip          |

**Example Request:**

```bash
curl "http://localhost:8085/api/v1/routes?gateway_id=gw_01HQ3K5V8N2M4P6R7S9T0UVW" \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "data": [
    {
      "id": "rt_01HQ3N9Y0Q5R7S9T1U3V4WXY",
      "gateway_id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
      "upstream_id": "ups_01HQ3M7X9P4Q6R8S0T2U3VWX",
      "path_prefix": "/api/users",
      "methods": ["GET", "POST", "PUT", "DELETE"],
      "strip_prefix": true,
      "priority": 10,
      "rate_limit_rps": 100,
      "created_at": "2026-02-20T10:45:00Z",
      "updated_at": "2026-02-20T10:45:00Z"
    }
  ],
  "pagination": {
    "total": 1,
    "limit": 20,
    "offset": 0
  }
}
```

---

#### `DELETE /api/v1/routes/{id}`

Delete a route by ID.

**Example Request:**

```bash
curl -X DELETE http://localhost:8085/api/v1/routes/rt_01HQ3N9Y0Q5R7S9T1U3V4WXY \
  -H "Authorization: Bearer <token>"
```

**Response `204 No Content`:**

No response body.

---

### Threats

#### `GET /api/v1/threats`

List detected threat events, optionally filtered by gateway.

**Query Parameters:**

| Parameter    | Type   | Required | Description            |
|------------- |--------|----------|------------------------|
| `gateway_id` | string | no       | Filter by gateway ID   |
| `limit`      | int    | no       | Items per page         |
| `offset`     | int    | no       | Items to skip          |

**Example Request:**

```bash
curl "http://localhost:8085/api/v1/threats?gateway_id=gw_01HQ3K5V8N2M4P6R7S9T0UVW" \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "data": [
    {
      "id": "thr_01HQ3PA10R6S8T0U2V4W5XYZ",
      "gateway_id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
      "threat_type": "QUANTUM_DOWNGRADE",
      "severity": "CRITICAL",
      "source_ip": "203.0.113.42",
      "description": "Client attempted to negotiate TLS 1.2 with RSA key exchange, bypassing PQC cipher suites",
      "details": {
        "requested_cipher": "TLS_RSA_WITH_AES_256_CBC_SHA256",
        "tls_version": "1.2",
        "gateway_policy": "PQC_PREFERRED"
      },
      "mitigated": false,
      "created_at": "2026-02-20T11:00:00Z"
    }
  ],
  "pagination": {
    "total": 1,
    "limit": 20,
    "offset": 0
  }
}
```

---

#### `GET /api/v1/threats/{id}`

Retrieve a single threat event by ID.

**Example Request:**

```bash
curl http://localhost:8085/api/v1/threats/thr_01HQ3PA10R6S8T0U2V4W5XYZ \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "id": "thr_01HQ3PA10R6S8T0U2V4W5XYZ",
  "gateway_id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
  "threat_type": "QUANTUM_DOWNGRADE",
  "severity": "CRITICAL",
  "source_ip": "203.0.113.42",
  "description": "Client attempted to negotiate TLS 1.2 with RSA key exchange, bypassing PQC cipher suites",
  "details": {
    "requested_cipher": "TLS_RSA_WITH_AES_256_CBC_SHA256",
    "tls_version": "1.2",
    "gateway_policy": "PQC_PREFERRED"
  },
  "mitigated": false,
  "created_at": "2026-02-20T11:00:00Z"
}
```

---

#### `POST /api/v1/threats/{id}/mitigate`

Mark a threat as mitigated and apply countermeasures.

**Example Request:**

```bash
curl -X POST http://localhost:8085/api/v1/threats/thr_01HQ3PA10R6S8T0U2V4W5XYZ/mitigate \
  -H "Authorization: Bearer <token>"
```

**Response `200 OK`:**

```json
{
  "id": "thr_01HQ3PA10R6S8T0U2V4W5XYZ",
  "mitigated": true,
  "mitigated_at": "2026-02-20T11:15:00Z",
  "message": "Threat mitigated successfully"
}
```

---

## AI Engine API (Port 8086)

Base URL: `http://localhost:8086`

### AI Engine Health

#### `GET /health`

Returns the health status of the AI engine. No authentication required.

**Response `200 OK`:**

```json
{
  "status": "healthy",
  "service": "qsgw-ai-engine",
  "version": "0.1.0",
  "models_loaded": true,
  "timestamp": "2026-02-20T10:30:00Z"
}
```

---

### Traffic Analysis

#### `POST /api/v1/analyze-traffic`

Analyze traffic patterns for anomalies and potential quantum-related threats.

**Request Body:**

| Field                    | Type   | Required | Description                                 |
|--------------------------|--------|----------|---------------------------------------------|
| `cipher_suite`           | string | yes      | TLS cipher suite used by the client         |
| `tls_version`            | string | yes      | TLS version (e.g., `1.3`, `1.2`)           |
| `handshake_duration_ms`  | float  | yes      | TLS handshake duration in milliseconds      |
| `requests_per_minute`    | float  | yes      | Current request rate from source IP         |
| `error_rate`             | float  | yes      | Error rate as a fraction (0.0 to 1.0)       |
| `unique_paths`           | int    | yes      | Number of unique URL paths accessed         |

**Example Request:**

```bash
curl -X POST http://localhost:8086/api/v1/analyze-traffic \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "cipher_suite": "TLS_RSA_WITH_AES_256_CBC_SHA256",
    "tls_version": "1.2",
    "handshake_duration_ms": 45.2,
    "requests_per_minute": 350.0,
    "error_rate": 0.15,
    "unique_paths": 87
  }'
```

**Response `200 OK`:**

```json
{
  "anomaly_score": 0.87,
  "is_anomalous": true,
  "threats": [
    {
      "threat_type": "QUANTUM_DOWNGRADE",
      "severity": "CRITICAL",
      "confidence": 0.92,
      "description": "Classical-only cipher suite detected; possible quantum downgrade attack"
    },
    {
      "threat_type": "ANOMALOUS_TRAFFIC",
      "severity": "MEDIUM",
      "confidence": 0.74,
      "description": "Request rate exceeds normal baseline by 3.5x"
    }
  ],
  "analysis": {
    "cipher_risk": "high",
    "tls_version_risk": "high",
    "rate_risk": "medium",
    "error_risk": "medium"
  }
}
```

---

### Bot Detection

#### `POST /api/v1/detect-bot`

Analyze request patterns to determine whether traffic originates from a bot.

**Request Body:**

| Field                   | Type   | Required | Description                                  |
|-------------------------|--------|----------|----------------------------------------------|
| `user_agent`            | string | yes      | HTTP User-Agent header value                 |
| `requests_per_minute`   | float  | yes      | Current request rate from source IP          |
| `unique_paths`          | int    | yes      | Number of unique URL paths accessed          |
| `error_rate`            | float  | yes      | Error rate as a fraction (0.0 to 1.0)        |
| `avg_response_time_ms`  | float  | yes      | Average response time in milliseconds        |

**Example Request:**

```bash
curl -X POST http://localhost:8086/api/v1/detect-bot \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_agent": "python-requests/2.31.0",
    "requests_per_minute": 500.0,
    "unique_paths": 200,
    "error_rate": 0.05,
    "avg_response_time_ms": 12.5
  }'
```

**Response `200 OK`:**

```json
{
  "is_bot": true,
  "bot_score": 0.91,
  "confidence": 0.88,
  "threat_type": "BOT_ATTACK",
  "severity": "HIGH",
  "indicators": [
    {
      "factor": "user_agent",
      "signal": "Known scripting library user-agent",
      "weight": 0.30
    },
    {
      "factor": "request_rate",
      "signal": "Request rate 5x above human baseline",
      "weight": 0.35
    },
    {
      "factor": "path_diversity",
      "signal": "High path diversity indicates automated crawling",
      "weight": 0.25
    },
    {
      "factor": "response_time",
      "signal": "Response processing time below human threshold",
      "weight": 0.10
    }
  ]
}
```

---

## Rate Limiting

All API endpoints are subject to rate limiting. When rate-limited, the API returns:

**Response `429 Too Many Requests`:**

```json
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Rate limit exceeded. Try again in 12 seconds.",
    "details": {
      "retry_after_seconds": 12
    }
  }
}
```

The `Retry-After` HTTP header is also included in the response.

---

## Complete Workflow Example

The following demonstrates a full workflow: creating a gateway, upstream, and route, then checking for threats.

```mermaid
sequenceDiagram
    participant Op as Operator
    participant CP as Control Plane<br/>:8085
    participant GW as Gateway<br/>:8443
    participant DB as PostgreSQL

    Op->>CP: POST /api/v1/gateways<br/>(name, hostname, port, tls_policy)
    CP->>DB: INSERT gateway
    DB-->>CP: gateway record
    CP-->>Op: 201 Created (gateway_id)

    Op->>CP: POST /api/v1/gateways/{id}/activate
    CP->>DB: UPDATE status = ACTIVE
    CP-->>Op: 200 OK (status: active)

    Op->>CP: POST /api/v1/upstreams<br/>(name, host, port, protocol)
    CP->>DB: INSERT upstream
    DB-->>CP: upstream record
    CP-->>Op: 201 Created (upstream_id)

    Op->>CP: POST /api/v1/routes<br/>(gateway_id, upstream_id, path_prefix)
    CP->>DB: INSERT route
    DB-->>CP: route record
    CP-->>Op: 201 Created (route_id)

    Note over GW: Gateway is now active,<br/>routing traffic to upstream

    Op->>CP: GET /api/v1/threats?gateway_id={id}
    CP->>DB: SELECT threat_events
    DB-->>CP: threat records
    CP-->>Op: 200 OK (threat list)
```

```bash
# 1. Create a gateway
curl -X POST http://localhost:8085/api/v1/gateways \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-gateway",
    "hostname": "0.0.0.0",
    "port": 8443,
    "tls_policy": "PQC_PREFERRED"
  }'

# 2. Activate the gateway
curl -X POST http://localhost:8085/api/v1/gateways/gw_01HQ3K5V8N2M4P6R7S9T0UVW/activate \
  -H "Authorization: Bearer $TOKEN"

# 3. Register an upstream
curl -X POST http://localhost:8085/api/v1/upstreams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "backend-api",
    "host": "10.0.1.100",
    "port": 8080,
    "protocol": "HTTP",
    "health_check_path": "/health"
  }'

# 4. Create a route
curl -X POST http://localhost:8085/api/v1/routes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "gateway_id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
    "upstream_id": "ups_01HQ3M7X9P4Q6R8S0T2U3VWX",
    "path_prefix": "/api",
    "methods": ["GET", "POST"],
    "strip_prefix": false,
    "priority": 10,
    "rate_limit_rps": 200
  }'

# 5. Monitor threats
curl "http://localhost:8085/api/v1/threats?gateway_id=gw_01HQ3K5V8N2M4P6R7S9T0UVW" \
  -H "Authorization: Bearer $TOKEN"
```
