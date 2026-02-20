-- QSGW Schema Rollback

DROP TRIGGER IF EXISTS trg_routes_updated_at ON routes;
DROP TRIGGER IF EXISTS trg_upstreams_updated_at ON upstreams;
DROP TRIGGER IF EXISTS trg_gateways_updated_at ON gateways;
DROP FUNCTION IF EXISTS qsgw_update_updated_at();

DROP TABLE IF EXISTS qsgw_audit_log;
DROP TABLE IF EXISTS threat_events;
DROP TABLE IF EXISTS tls_sessions;
DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS upstreams;
DROP TABLE IF EXISTS gateways;

DROP TYPE IF EXISTS threat_type;
DROP TYPE IF EXISTS threat_severity;
DROP TYPE IF EXISTS tls_policy;
DROP TYPE IF EXISTS route_protocol;
DROP TYPE IF EXISTS gateway_status;
