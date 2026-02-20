# Deployment Guide

This guide covers production deployment of QSGW, including Docker Compose configuration, Kubernetes manifests, environment variables, database setup, and operational best practices.

## Table of Contents

- [Docker Compose Production Setup](#docker-compose-production-setup)
- [Environment Variable Reference](#environment-variable-reference)
- [PostgreSQL Setup](#postgresql-setup)
- [etcd Cluster Setup](#etcd-cluster-setup)
- [TLS Certificate Provisioning](#tls-certificate-provisioning)
- [Monitoring](#monitoring)
- [High Availability](#high-availability)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Scaling Recommendations](#scaling-recommendations)
- [Backup and Recovery](#backup-and-recovery)
- [Security Hardening Checklist](#security-hardening-checklist)

---

## Docker Compose Production Setup

The following Docker Compose configuration is suitable for production deployments with resource limits, health checks, and proper networking.

```yaml
version: "3.9"

services:
  gateway:
    image: quantun/qsgw-gateway:0.1.0
    ports:
      - "8443:8443"
    environment:
      - QSGW_GATEWAY_HOST=0.0.0.0
      - QSGW_GATEWAY_PORT=8443
      - QSGW_TLS_CERT_PATH=/etc/qsgw/certs/server.crt
      - QSGW_TLS_KEY_PATH=/etc/qsgw/certs/server.key
      - QSGW_CONTROL_PLANE_URL=http://control-plane:8085
      - QSGW_AI_ENGINE_URL=http://ai-engine:8086
      - QSGW_ETCD_ENDPOINTS=http://etcd:2379
      - QSGW_MAX_CONNECTIONS=50000
      - QSGW_WORKER_THREADS=8
      - QSGW_LOG_LEVEL=info
    volumes:
      - ./certs:/etc/qsgw/certs:ro
    deploy:
      resources:
        limits:
          cpus: "4.0"
          memory: 2G
        reservations:
          cpus: "2.0"
          memory: 1G
    restart: unless-stopped
    depends_on:
      control-plane:
        condition: service_healthy
    networks:
      - qsgw-internal

  control-plane:
    image: quantun/qsgw-control-plane:0.1.0
    ports:
      - "8085:8085"
    environment:
      - QSGW_CP_HOST=0.0.0.0
      - QSGW_CP_PORT=8085
      - QSGW_DATABASE_URL=postgres://qsgw:${QSGW_DB_PASSWORD}@postgres:5432/qsgw?sslmode=require
      - QSGW_DATABASE_MAX_CONNS=50
      - QSGW_DATABASE_MIN_CONNS=10
      - QSGW_JWT_SECRET=${QSGW_JWT_SECRET}
      - QSGW_API_KEY=${QSGW_API_KEY}
      - QSGW_ETCD_ENDPOINTS=http://etcd:2379
      - QSGW_LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8085/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 15s
    deploy:
      resources:
        limits:
          cpus: "2.0"
          memory: 1G
        reservations:
          cpus: "1.0"
          memory: 512M
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      etcd:
        condition: service_healthy
    networks:
      - qsgw-internal

  ai-engine:
    image: quantun/qsgw-ai-engine:0.1.0
    ports:
      - "8086:8086"
    environment:
      - QSGW_AI_HOST=0.0.0.0
      - QSGW_AI_PORT=8086
      - QSGW_AI_WORKERS=4
      - QSGW_ANOMALY_THRESHOLD=0.65
      - QSGW_BOT_THRESHOLD=0.70
      - QSGW_LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "python", "-c", "import urllib.request; urllib.request.urlopen('http://localhost:8086/health')"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 20s
    deploy:
      resources:
        limits:
          cpus: "2.0"
          memory: 1G
        reservations:
          cpus: "1.0"
          memory: 512M
    restart: unless-stopped
    networks:
      - qsgw-internal

  admin:
    image: quantun/qsgw-admin:0.1.0
    ports:
      - "3003:3003"
    environment:
      - VITE_API_URL=http://localhost:8085
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 256M
        reservations:
          cpus: "0.25"
          memory: 128M
    restart: unless-stopped
    networks:
      - qsgw-internal

  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_DB=qsgw
      - POSTGRES_USER=qsgw
      - POSTGRES_PASSWORD=${QSGW_DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U qsgw -d qsgw"]
      interval: 5s
      timeout: 3s
      retries: 5
    deploy:
      resources:
        limits:
          cpus: "2.0"
          memory: 2G
        reservations:
          cpus: "1.0"
          memory: 1G
    restart: unless-stopped
    networks:
      - qsgw-internal

  etcd:
    image: quay.io/coreos/etcd:v3.5.12
    command:
      - etcd
      - --name=etcd0
      - --data-dir=/etcd-data
      - --listen-client-urls=http://0.0.0.0:2379
      - --advertise-client-urls=http://etcd:2379
      - --listen-peer-urls=http://0.0.0.0:2380
    volumes:
      - etcddata:/etcd-data
    healthcheck:
      test: ["CMD", "etcdctl", "endpoint", "health"]
      interval: 10s
      timeout: 5s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: "1.0"
          memory: 512M
        reservations:
          cpus: "0.5"
          memory: 256M
    restart: unless-stopped
    networks:
      - qsgw-internal

volumes:
  pgdata:
  etcddata:

networks:
  qsgw-internal:
    driver: bridge
```

**Start the stack:**

```bash
# Create a .env file with secrets
cat > .env << 'EOF'
QSGW_DB_PASSWORD=your-secure-database-password
QSGW_JWT_SECRET=your-256-bit-jwt-secret
QSGW_API_KEY=qsgw_k_your-api-key
EOF

docker compose -f docker-compose.prod.yml up -d
```

---

## Environment Variable Reference

### Gateway (Rust)

| Variable                     | Default            | Description                              |
|------------------------------|--------------------|------------------------------------------|
| `QSGW_GATEWAY_HOST`         | `0.0.0.0`         | Bind address                             |
| `QSGW_GATEWAY_PORT`         | `8443`             | Listen port                              |
| `QSGW_TLS_CERT_PATH`        | `/etc/qsgw/certs/server.crt` | TLS certificate path         |
| `QSGW_TLS_KEY_PATH`         | `/etc/qsgw/certs/server.key` | TLS private key path         |
| `QSGW_CONTROL_PLANE_URL`    | `http://localhost:8085` | Control plane URL                   |
| `QSGW_AI_ENGINE_URL`        | `http://localhost:8086` | AI engine URL                       |
| `QSGW_ETCD_ENDPOINTS`       | `http://localhost:2379` | etcd connection endpoints           |
| `QSGW_MAX_CONNECTIONS`      | `10000`            | Maximum concurrent connections           |
| `QSGW_WORKER_THREADS`       | CPU cores          | Tokio worker thread count                |
| `QSGW_BLOCKING_THREADS`     | `512`              | Maximum blocking thread pool             |
| `QSGW_UPSTREAM_POOL_SIZE`   | `100`              | Connection pool size per upstream        |
| `QSGW_UPSTREAM_TIMEOUT_SECS`| `30`               | Upstream request timeout                 |
| `QSGW_RATE_LIMIT_RPS`       | `100`              | Per-IP rate limit (requests/sec)         |
| `QSGW_RATE_LIMIT_BURST`     | `200`              | Rate limit burst size                    |
| `QSGW_LOG_LEVEL`            | `info`             | Log level (trace, debug, info, warn, error) |

### Control Plane (Go)

| Variable                     | Default            | Description                              |
|------------------------------|--------------------|------------------------------------------|
| `QSGW_CP_HOST`              | `0.0.0.0`         | Bind address                             |
| `QSGW_CP_PORT`              | `8085`             | Listen port                              |
| `QSGW_DATABASE_URL`         | --                 | PostgreSQL connection string (required)  |
| `QSGW_DATABASE_MAX_CONNS`   | `25`               | Maximum database connections             |
| `QSGW_DATABASE_MIN_CONNS`   | `5`                | Minimum idle database connections        |
| `QSGW_JWT_SECRET`           | --                 | JWT signing secret (required)            |
| `QSGW_API_KEY`              | --                 | API key for service authentication       |
| `QSGW_ETCD_ENDPOINTS`       | `http://localhost:2379` | etcd connection endpoints           |
| `QSGW_WEBHOOK_URL`          | --                 | Webhook endpoint for threat alerts       |
| `QSGW_WEBHOOK_SECRET`       | --                 | HMAC-SHA256 secret for webhook signing   |
| `QSGW_WEBHOOK_MIN_SEVERITY` | `MEDIUM`           | Minimum severity for webhook alerts      |
| `QSGW_LOG_LEVEL`            | `info`             | Log level (debug, info, warn, error)     |

### AI Engine (Python)

| Variable                          | Default   | Description                             |
|-----------------------------------|-----------|-----------------------------------------|
| `QSGW_AI_HOST`                   | `0.0.0.0` | Bind address                           |
| `QSGW_AI_PORT`                   | `8086`    | Listen port                             |
| `QSGW_AI_WORKERS`                | `4`       | Uvicorn worker processes                |
| `QSGW_ANOMALY_THRESHOLD`         | `0.65`    | Anomaly detection score threshold       |
| `QSGW_ANOMALY_CRITICAL_THRESHOLD`| `0.85`    | Critical anomaly threshold              |
| `QSGW_ANOMALY_RATE_THRESHOLD`    | `300`     | RPM threshold for rate anomalies        |
| `QSGW_ANOMALY_RATE_MULTIPLIER`   | `3.0`     | Baseline multiplier for rate alerts     |
| `QSGW_BOT_THRESHOLD`             | `0.70`    | Bot detection score threshold           |
| `QSGW_BOT_BLOCK_THRESHOLD`       | `0.90`    | Bot blocking score threshold            |
| `QSGW_LOG_LEVEL`                 | `info`    | Log level (debug, info, warning, error) |

### Admin Dashboard (React)

| Variable          | Default                  | Description                   |
|-------------------|--------------------------|-------------------------------|
| `VITE_API_URL`    | `http://localhost:8085`  | Control plane API URL         |

---

## PostgreSQL Setup

### Connection Pooling

The control plane uses pgx v5 with built-in connection pooling. Configure pool sizes based on your workload:

```bash
# For moderate traffic (< 1,000 rps)
export QSGW_DATABASE_MAX_CONNS=25
export QSGW_DATABASE_MIN_CONNS=5

# For high traffic (> 5,000 rps)
export QSGW_DATABASE_MAX_CONNS=100
export QSGW_DATABASE_MIN_CONNS=25
```

### Database Migrations

Migrations are applied automatically on first startup. To run them manually:

```bash
# Apply migrations
psql -h localhost -U qsgw -d qsgw -f migrations/001_initial_schema.sql
```

The schema creates the following tables:

- `gateways` -- Gateway instance configuration
- `upstreams` -- Backend service definitions
- `routes` -- Routing rules mapping paths to upstreams
- `tls_sessions` -- TLS session metadata for analysis
- `threat_events` -- Detected threat events
- `qsgw_audit_log` -- Administrative action audit trail

### Backups

Schedule regular backups using `pg_dump`:

```bash
# Full database backup
pg_dump -h localhost -U qsgw -d qsgw -Fc -f qsgw_backup_$(date +%Y%m%d).dump

# Restore from backup
pg_restore -h localhost -U qsgw -d qsgw -c qsgw_backup_20260220.dump
```

For production, configure continuous archiving with WAL shipping for point-in-time recovery.

### PostgreSQL Tuning

Recommended `postgresql.conf` settings for QSGW workloads:

```ini
# Connection handling
max_connections = 200
shared_buffers = 1GB
effective_cache_size = 3GB
work_mem = 16MB

# Write performance
wal_buffers = 64MB
checkpoint_completion_target = 0.9
synchronous_commit = on

# Query planning
random_page_cost = 1.1
effective_io_concurrency = 200
```

---

## etcd Cluster Setup

QSGW uses etcd for distributed configuration. Gateway instances watch etcd keys for real-time configuration updates.

### Single Node (Development)

```bash
etcd \
  --name=etcd0 \
  --data-dir=/var/lib/etcd \
  --listen-client-urls=http://0.0.0.0:2379 \
  --advertise-client-urls=http://etcd:2379
```

### Three-Node Cluster (Production)

For high availability, deploy a three-node etcd cluster:

```bash
# Node 1
etcd \
  --name=etcd1 \
  --initial-advertise-peer-urls=http://etcd1:2380 \
  --listen-peer-urls=http://0.0.0.0:2380 \
  --advertise-client-urls=http://etcd1:2379 \
  --listen-client-urls=http://0.0.0.0:2379 \
  --initial-cluster=etcd1=http://etcd1:2380,etcd2=http://etcd2:2380,etcd3=http://etcd3:2380 \
  --initial-cluster-state=new

# Node 2
etcd \
  --name=etcd2 \
  --initial-advertise-peer-urls=http://etcd2:2380 \
  --listen-peer-urls=http://0.0.0.0:2380 \
  --advertise-client-urls=http://etcd2:2379 \
  --listen-client-urls=http://0.0.0.0:2379 \
  --initial-cluster=etcd1=http://etcd1:2380,etcd2=http://etcd2:2380,etcd3=http://etcd3:2380 \
  --initial-cluster-state=new

# Node 3
etcd \
  --name=etcd3 \
  --initial-advertise-peer-urls=http://etcd3:2380 \
  --listen-peer-urls=http://0.0.0.0:2380 \
  --advertise-client-urls=http://etcd3:2379 \
  --listen-client-urls=http://0.0.0.0:2379 \
  --initial-cluster=etcd1=http://etcd1:2380,etcd2=http://etcd2:2380,etcd3=http://etcd3:2380 \
  --initial-cluster-state=new
```

Set the gateway's etcd endpoints to include all cluster members:

```bash
export QSGW_ETCD_ENDPOINTS=http://etcd1:2379,http://etcd2:2379,http://etcd3:2379
```

---

## TLS Certificate Provisioning

### Self-Signed Certificates (Development)

```bash
# Generate a self-signed certificate for development
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
  -days 365 -nodes -subj "/CN=localhost"
```

### Production Certificates

For production deployments, obtain certificates from a trusted Certificate Authority. Place the certificate chain and private key at the paths configured in `QSGW_TLS_CERT_PATH` and `QSGW_TLS_KEY_PATH`.

**Using Let's Encrypt with certbot:**

```bash
certbot certonly --standalone -d gateway.example.com
cp /etc/letsencrypt/live/gateway.example.com/fullchain.pem /etc/qsgw/certs/server.crt
cp /etc/letsencrypt/live/gateway.example.com/privkey.pem /etc/qsgw/certs/server.key
```

### PQC Certificates

For full post-quantum TLS, generate certificates using ML-DSA or hybrid algorithms with a PQC-capable CA. The QSGW crypto crate includes utilities for generating PQC key pairs and certificate signing requests.

---

## Monitoring

### Health Endpoints

Each service exposes a health endpoint that returns `200 OK` when the service is operational:

| Service        | Endpoint                        |
|----------------|---------------------------------|
| Gateway        | `https://localhost:8443/health` |
| Control Plane  | `http://localhost:8085/health`  |
| AI Engine      | `http://localhost:8086/health`  |

### Structured Logging

All services emit structured JSON logs for integration with log aggregation systems.

**Gateway (Rust):** Uses `tracing` with JSON output.

**Control Plane (Go):** Uses `zap` with structured JSON fields.

**AI Engine (Python):** Uses `structlog` with JSON rendering.

Example log entry:

```json
{
  "level": "info",
  "ts": "2026-02-20T10:30:00.000Z",
  "caller": "server/main.go:85",
  "msg": "request completed",
  "method": "GET",
  "path": "/api/v1/gateways",
  "status": 200,
  "duration_ms": 12.5,
  "request_id": "req_abc123"
}
```

### Metrics

Expose Prometheus-compatible metrics by setting `QSGW_METRICS_ENABLED=true`:

| Metric                           | Type      | Description                      |
|----------------------------------|-----------|----------------------------------|
| `qsgw_connections_total`         | Counter   | Total connections received       |
| `qsgw_connections_active`        | Gauge     | Current active connections       |
| `qsgw_requests_total`            | Counter   | Total proxied requests           |
| `qsgw_request_duration_seconds`  | Histogram | Request processing duration      |
| `qsgw_threats_total`             | Counter   | Total threat events by type      |
| `qsgw_tls_handshake_duration_seconds` | Histogram | TLS handshake latency      |
| `qsgw_upstream_health`           | Gauge     | Upstream health status (0/1)     |

---

## High Availability

### Multiple Gateway Instances

Deploy multiple gateway instances behind a Layer 4 (TCP) load balancer for high availability and horizontal scaling.

```
              [L4 Load Balancer]
              /        |        \
        [Gateway 1] [Gateway 2] [Gateway 3]
              \        |        /
           [Control Plane Cluster]
                      |
               [PostgreSQL HA]
```

Each gateway instance connects to the same control plane and etcd cluster. Route configuration is synchronized through etcd watches.

### etcd Cluster

Run a minimum of 3 etcd nodes for fault tolerance. The cluster tolerates the loss of 1 node (Raft consensus requires a majority).

| Cluster Size | Fault Tolerance |
|------------- |-----------------|
| 1 node       | 0 failures      |
| 3 nodes      | 1 failure       |
| 5 nodes      | 2 failures      |

### PostgreSQL Replication

For database high availability, configure PostgreSQL streaming replication:

1. **Primary:** Handles all writes from the control plane.
2. **Replica(s):** Provide read-only copies for reporting and failover.
3. **Failover:** Use Patroni or pg_auto_failover for automatic promotion.

---

## Kubernetes Deployment

### Gateway Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: qsgw-gateway
  labels:
    app: qsgw
    component: gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: qsgw
      component: gateway
  template:
    metadata:
      labels:
        app: qsgw
        component: gateway
    spec:
      containers:
        - name: gateway
          image: quantun/qsgw-gateway:0.1.0
          ports:
            - containerPort: 8443
              protocol: TCP
          envFrom:
            - configMapRef:
                name: qsgw-config
            - secretRef:
                name: qsgw-secrets
          resources:
            requests:
              cpu: "2"
              memory: 1Gi
            limits:
              cpu: "4"
              memory: 2Gi
          volumeMounts:
            - name: tls-certs
              mountPath: /etc/qsgw/certs
              readOnly: true
      volumes:
        - name: tls-certs
          secret:
            secretName: qsgw-tls-cert
---
apiVersion: v1
kind: Service
metadata:
  name: qsgw-gateway
spec:
  type: LoadBalancer
  selector:
    app: qsgw
    component: gateway
  ports:
    - port: 8443
      targetPort: 8443
      protocol: TCP
```

### Control Plane Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: qsgw-control-plane
  labels:
    app: qsgw
    component: control-plane
spec:
  replicas: 2
  selector:
    matchLabels:
      app: qsgw
      component: control-plane
  template:
    metadata:
      labels:
        app: qsgw
        component: control-plane
    spec:
      containers:
        - name: control-plane
          image: quantun/qsgw-control-plane:0.1.0
          ports:
            - containerPort: 8085
          envFrom:
            - configMapRef:
                name: qsgw-config
            - secretRef:
                name: qsgw-secrets
          resources:
            requests:
              cpu: "1"
              memory: 512Mi
            limits:
              cpu: "2"
              memory: 1Gi
          livenessProbe:
            httpGet:
              path: /health
              port: 8085
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8085
            initialDelaySeconds: 5
            periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: qsgw-control-plane
spec:
  type: ClusterIP
  selector:
    app: qsgw
    component: control-plane
  ports:
    - port: 8085
      targetPort: 8085
```

### AI Engine Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: qsgw-ai-engine
  labels:
    app: qsgw
    component: ai-engine
spec:
  replicas: 2
  selector:
    matchLabels:
      app: qsgw
      component: ai-engine
  template:
    metadata:
      labels:
        app: qsgw
        component: ai-engine
    spec:
      containers:
        - name: ai-engine
          image: quantun/qsgw-ai-engine:0.1.0
          ports:
            - containerPort: 8086
          envFrom:
            - configMapRef:
                name: qsgw-config
          resources:
            requests:
              cpu: "1"
              memory: 512Mi
            limits:
              cpu: "2"
              memory: 1Gi
          livenessProbe:
            httpGet:
              path: /health
              port: 8086
            initialDelaySeconds: 15
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8086
            initialDelaySeconds: 10
            periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: qsgw-ai-engine
spec:
  type: ClusterIP
  selector:
    app: qsgw
    component: ai-engine
  ports:
    - port: 8086
      targetPort: 8086
```

### Admin Dashboard Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: qsgw-admin
  labels:
    app: qsgw
    component: admin
spec:
  replicas: 1
  selector:
    matchLabels:
      app: qsgw
      component: admin
  template:
    metadata:
      labels:
        app: qsgw
        component: admin
    spec:
      containers:
        - name: admin
          image: quantun/qsgw-admin:0.1.0
          ports:
            - containerPort: 3003
          envFrom:
            - configMapRef:
                name: qsgw-config
          resources:
            requests:
              cpu: "250m"
              memory: 128Mi
            limits:
              cpu: "500m"
              memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: qsgw-admin
spec:
  type: ClusterIP
  selector:
    app: qsgw
    component: admin
  ports:
    - port: 3003
      targetPort: 3003
```

### ConfigMap and Secrets

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: qsgw-config
data:
  QSGW_GATEWAY_HOST: "0.0.0.0"
  QSGW_GATEWAY_PORT: "8443"
  QSGW_CP_HOST: "0.0.0.0"
  QSGW_CP_PORT: "8085"
  QSGW_AI_HOST: "0.0.0.0"
  QSGW_AI_PORT: "8086"
  QSGW_CONTROL_PLANE_URL: "http://qsgw-control-plane:8085"
  QSGW_AI_ENGINE_URL: "http://qsgw-ai-engine:8086"
  QSGW_ETCD_ENDPOINTS: "http://etcd:2379"
  QSGW_MAX_CONNECTIONS: "50000"
  QSGW_LOG_LEVEL: "info"
  QSGW_ANOMALY_THRESHOLD: "0.65"
  QSGW_BOT_THRESHOLD: "0.70"
  VITE_API_URL: "http://qsgw-control-plane:8085"
---
apiVersion: v1
kind: Secret
metadata:
  name: qsgw-secrets
type: Opaque
stringData:
  QSGW_DATABASE_URL: "postgres://qsgw:password@postgres:5432/qsgw?sslmode=require"
  QSGW_JWT_SECRET: "your-256-bit-jwt-secret"
  QSGW_API_KEY: "qsgw_k_your-api-key"
---
apiVersion: v1
kind: Secret
metadata:
  name: qsgw-tls-cert
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-certificate>
  tls.key: <base64-encoded-private-key>
```

---

## Scaling Recommendations

| Traffic Level | Gateway Replicas | Control Plane Replicas | AI Engine Replicas | PostgreSQL | etcd Nodes |
|-------------- |------------------|------------------------|--------------------|-----------  |------------|
| Low (< 1K rps)      | 1           | 1                      | 1                  | Single     | 1          |
| Medium (1--10K rps)  | 2--3        | 2                      | 2                  | Primary + 1 replica | 3 |
| High (10--50K rps)   | 4--8        | 3                      | 3--4               | Primary + 2 replicas | 3 |
| Very High (50K+ rps) | 8--16       | 4                      | 4--8               | Primary + 2 replicas + pgBouncer | 5 |

**Scaling guidelines:**

- **Gateway:** Scale horizontally. Each instance handles up to 50,000 concurrent connections. Add instances behind the load balancer as needed.
- **Control Plane:** Stateless; scale horizontally. Database connections are the primary bottleneck.
- **AI Engine:** CPU-bound. Scale based on analysis request volume. Each worker process handles approximately 2,000 analysis requests per second.
- **PostgreSQL:** Scale vertically first (more CPU, memory, faster storage). Add read replicas for reporting workloads.

---

## Backup and Recovery

### Database Backups

**Automated daily backups:**

```bash
#!/bin/bash
# /etc/cron.daily/qsgw-backup
BACKUP_DIR=/var/backups/qsgw
DATE=$(date +%Y%m%d_%H%M%S)

pg_dump -h localhost -U qsgw -d qsgw -Fc -f "$BACKUP_DIR/qsgw_$DATE.dump"

# Retain 30 days of backups
find "$BACKUP_DIR" -name "qsgw_*.dump" -mtime +30 -delete
```

### etcd Backups

```bash
# Snapshot the etcd cluster
etcdctl snapshot save /var/backups/etcd/snapshot_$(date +%Y%m%d).db

# Restore from snapshot
etcdctl snapshot restore /var/backups/etcd/snapshot_20260220.db \
  --data-dir=/var/lib/etcd-restore
```

### Recovery Procedures

1. **Database failure:** Restore from the latest `pg_dump` backup. Apply WAL archives for point-in-time recovery if continuous archiving is configured.
2. **etcd failure:** Restore from the latest snapshot. Gateway instances will reconnect and resynchronize configuration.
3. **Gateway failure:** The load balancer routes traffic to healthy instances. Replace the failed instance and it will pull configuration from etcd.

---

## Security Hardening Checklist

- [ ] **TLS certificates:** Use certificates from a trusted CA. Rotate certificates before expiration.
- [ ] **Secrets management:** Store `QSGW_JWT_SECRET`, `QSGW_API_KEY`, and `QSGW_DB_PASSWORD` in a secrets manager (Vault, AWS Secrets Manager, Kubernetes Secrets). Never commit secrets to version control.
- [ ] **Network segmentation:** Place the gateway in a DMZ. Keep the control plane, AI engine, PostgreSQL, and etcd on internal networks.
- [ ] **Firewall rules:** Only expose port 8443 (gateway) externally. Restrict ports 8085, 8086, 3003, 5432, and 2379 to internal networks.
- [ ] **Database encryption:** Enable SSL/TLS for PostgreSQL connections (`sslmode=require`).
- [ ] **etcd authentication:** Enable etcd client certificate authentication for production clusters.
- [ ] **Resource limits:** Set CPU and memory limits on all containers to prevent resource exhaustion.
- [ ] **Log retention:** Configure log rotation and retention policies. Ship logs to a centralized system.
- [ ] **Vulnerability scanning:** Regularly scan container images for known vulnerabilities.
- [ ] **Updates:** Monitor QSGW releases and apply security patches promptly.
- [ ] **TLS policy:** Use `PQC_PREFERRED` or `PQC_ONLY` in production. Avoid `CLASSICAL_ALLOWED` unless required for legacy compatibility.
- [ ] **Rate limiting:** Configure appropriate per-IP and per-route rate limits to prevent abuse.
- [ ] **Audit logging:** Verify that the `qsgw_audit_log` table captures all administrative actions.
