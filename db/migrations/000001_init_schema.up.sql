-- QSGW (Quantum-Safe Gateway) Schema

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enum types
CREATE TYPE gateway_status AS ENUM ('ACTIVE', 'INACTIVE', 'DRAINING', 'FAILED');
CREATE TYPE route_protocol AS ENUM ('HTTP', 'HTTPS', 'GRPC', 'TCP', 'TLS');
CREATE TYPE tls_policy AS ENUM ('PQC_ONLY', 'PQC_PREFERRED', 'HYBRID', 'CLASSICAL_ALLOWED');
CREATE TYPE threat_severity AS ENUM ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO');
CREATE TYPE threat_type AS ENUM ('QUANTUM_DOWNGRADE', 'WEAK_CIPHER', 'BOT_ATTACK', 'ANOMALOUS_TRAFFIC', 'CERTIFICATE_ISSUE', 'REPLAY_ATTACK');

-- Gateway instances
CREATE TABLE gateways (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    hostname VARCHAR(512) NOT NULL,
    port INTEGER NOT NULL DEFAULT 443,
    status gateway_status NOT NULL DEFAULT 'INACTIVE',
    tls_policy tls_policy NOT NULL DEFAULT 'PQC_PREFERRED',
    tls_cert_path VARCHAR(1024),
    tls_key_path VARCHAR(1024),
    max_connections INTEGER NOT NULL DEFAULT 10000,
    metadata JSONB DEFAULT '{}',
    created_by VARCHAR(255) NOT NULL DEFAULT 'system',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by VARCHAR(255) NOT NULL DEFAULT 'system',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Upstream services
CREATE TABLE upstreams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    host VARCHAR(512) NOT NULL,
    port INTEGER NOT NULL,
    protocol route_protocol NOT NULL DEFAULT 'HTTPS',
    tls_verify BOOLEAN NOT NULL DEFAULT true,
    health_check_path VARCHAR(512) DEFAULT '/health',
    health_check_interval_secs INTEGER NOT NULL DEFAULT 30,
    is_healthy BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Routes map incoming requests to upstreams
CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    gateway_id UUID NOT NULL REFERENCES gateways(id) ON DELETE CASCADE,
    upstream_id UUID NOT NULL REFERENCES upstreams(id) ON DELETE CASCADE,
    path_prefix VARCHAR(1024) NOT NULL,
    methods VARCHAR(255)[] DEFAULT '{}',
    strip_prefix BOOLEAN NOT NULL DEFAULT false,
    priority INTEGER NOT NULL DEFAULT 100,
    tls_policy tls_policy,
    rate_limit_rps INTEGER,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(gateway_id, path_prefix)
);

-- TLS sessions for PQC handshake tracking
CREATE TABLE tls_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    gateway_id UUID NOT NULL REFERENCES gateways(id) ON DELETE CASCADE,
    client_ip VARCHAR(45) NOT NULL,
    cipher_suite VARCHAR(255) NOT NULL,
    tls_version VARCHAR(10) NOT NULL,
    kem_algorithm VARCHAR(100),
    sig_algorithm VARCHAR(100),
    is_pqc BOOLEAN NOT NULL DEFAULT false,
    handshake_duration_ms INTEGER,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

-- Threat events detected by AI engine
CREATE TABLE threat_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    gateway_id UUID NOT NULL REFERENCES gateways(id) ON DELETE CASCADE,
    threat_type threat_type NOT NULL,
    severity threat_severity NOT NULL,
    source_ip VARCHAR(45),
    description TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    mitigated BOOLEAN NOT NULL DEFAULT false,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit log
CREATE TABLE qsgw_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor VARCHAR(255) NOT NULL DEFAULT 'system',
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_routes_gateway ON routes(gateway_id);
CREATE INDEX idx_tls_sessions_gateway ON tls_sessions(gateway_id);
CREATE INDEX idx_tls_sessions_started ON tls_sessions(started_at);
CREATE INDEX idx_threat_events_gateway ON threat_events(gateway_id);
CREATE INDEX idx_threat_events_detected ON threat_events(detected_at);
CREATE INDEX idx_threat_events_severity ON threat_events(severity);
CREATE INDEX idx_qsgw_audit_log_entity ON qsgw_audit_log(entity_type, entity_id);

-- Update trigger
CREATE OR REPLACE FUNCTION qsgw_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_gateways_updated_at BEFORE UPDATE ON gateways
    FOR EACH ROW EXECUTE FUNCTION qsgw_update_updated_at();

CREATE TRIGGER trg_upstreams_updated_at BEFORE UPDATE ON upstreams
    FOR EACH ROW EXECUTE FUNCTION qsgw_update_updated_at();

CREATE TRIGGER trg_routes_updated_at BEFORE UPDATE ON routes
    FOR EACH ROW EXECUTE FUNCTION qsgw_update_updated_at();
