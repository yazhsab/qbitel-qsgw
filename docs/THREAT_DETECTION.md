# Threat Detection Guide

QSGW integrates an AI-powered threat analysis pipeline that monitors all proxied connections in real time. This guide covers the detection capabilities, threat taxonomy, scoring models, mitigation workflows, and extensibility of the system.

## Table of Contents

- [Overview](#overview)
- [Anomaly Detection](#anomaly-detection)
- [Bot Detection](#bot-detection)
- [Threat Types Reference](#threat-types-reference)
- [Severity Levels](#severity-levels)
- [Threat Mitigation](#threat-mitigation)
- [Integration with the Gateway](#integration-with-the-gateway)
- [Extending Detection](#extending-detection)
- [Alert Configuration](#alert-configuration)

---

## Overview

The QSGW threat detection system operates as a dedicated AI engine (Python, FastAPI) running on port 8086. It receives telemetry from the Rust gateway engine and the Go control plane, processes it through machine learning models, and returns threat assessments in real time.

**Architecture:**

```
Client Connection
       |
       v
  [Rust Gateway]  --- TLS session metadata --->  [AI Engine]
       |                                              |
       v                                              v
  [Proxy traffic]                              [Anomaly Detector]
                                               [Bot Detector]
                                                      |
                                                      v
                                              [Threat Events]
                                                      |
                                                      v
                                            [Control Plane DB]
                                            [Admin Dashboard]
```

**Key properties:**

- Real-time analysis with sub-millisecond scoring latency
- Two independent detection models: anomaly detection and bot detection
- Composite scoring with configurable thresholds
- All threat events are persisted in PostgreSQL for audit and investigation
- Threat events are surfaced in the admin dashboard and via the REST API

---

## Anomaly Detection

The anomaly detector analyzes TLS session metadata and traffic patterns to identify connections that deviate from normal behavior. It is specifically designed to catch quantum-related attacks that would bypass traditional threat detection systems.

### Input Features

The anomaly detector evaluates the following features for each analyzed connection:

| Feature                  | Type   | Description                                      |
|--------------------------|--------|--------------------------------------------------|
| `cipher_suite`           | string | Negotiated TLS cipher suite                      |
| `tls_version`            | string | Negotiated TLS version                           |
| `handshake_duration_ms`  | float  | TLS handshake completion time in milliseconds    |
| `requests_per_minute`    | float  | Request rate from the source IP                  |
| `error_rate`             | float  | Fraction of requests resulting in errors (0--1)  |
| `unique_paths`           | int    | Number of distinct URL paths accessed            |

### Detection Categories

#### Quantum Downgrade Detection

Monitors for clients that request or negotiate classical-only cipher suites when the gateway policy requires post-quantum protection. This is the primary defense against harvest-now-decrypt-later strategies.

**Indicators:**

- Client offers only RSA, ECDHE, or other classical key exchange mechanisms
- TLS version negotiated below 1.3 when PQC requires 1.3
- Cipher suite does not include ML-KEM or hybrid algorithms
- Repeated reconnection attempts with progressively weaker cipher proposals

**Trigger conditions:**

- Gateway TLS policy is `PQC_ONLY` or `PQC_PREFERRED` and the client negotiates a classical-only suite
- Gateway TLS policy is `HYBRID` and the client negotiates without any PQC component

#### Traffic Spike Detection

Identifies abnormal surges in request volume that may indicate DDoS attacks, credential stuffing, or automated scanning.

**Indicators:**

- `requests_per_minute` exceeds the rolling baseline by more than 3x
- Sudden increase from a previously quiet source IP
- Coordinated spikes from multiple source IPs in the same subnet

**Thresholds (configurable):**

| Parameter                     | Default | Description                          |
|-------------------------------|---------|--------------------------------------|
| `QSGW_ANOMALY_RATE_THRESHOLD` | 300     | RPM above which traffic is flagged   |
| `QSGW_ANOMALY_RATE_MULTIPLIER`| 3.0     | Multiplier over baseline for alerts  |

#### Handshake Anomaly Detection

Unusual TLS handshake durations can indicate protocol manipulation, side-channel probing, or resource exhaustion attacks.

**Indicators:**

- Handshake duration significantly above the 99th percentile (>500ms for PQC, >200ms for classical)
- Handshake duration near zero (possible replay or pre-computed attack)
- High variance in handshake times from the same source IP

#### Error Rate Monitoring

Elevated error rates from a single source may indicate brute-force attacks, fuzzing, or path traversal attempts.

**Indicators:**

- Error rate above 0.10 (10%) from a single IP
- Sustained error rates above 0.05 over a 5-minute window
- Error codes concentrated on 401/403 (credential attacks) or 404 (path scanning)

### Scoring Model

The anomaly detector produces a composite anomaly score between 0.0 (normal) and 1.0 (highly anomalous). The score is computed as a weighted sum of individual risk signals:

| Signal              | Weight | Description                              |
|---------------------|--------|------------------------------------------|
| Cipher risk         | 0.35   | Classical or weak cipher suite penalty   |
| TLS version risk    | 0.20   | Sub-1.3 TLS version penalty              |
| Rate risk           | 0.20   | Requests per minute above threshold      |
| Error risk          | 0.15   | Elevated error rate penalty              |
| Handshake risk      | 0.10   | Unusual handshake duration penalty       |

**Threshold configuration:**

| Parameter                      | Default | Description                        |
|--------------------------------|---------|------------------------------------|
| `QSGW_ANOMALY_THRESHOLD`      | 0.65    | Score above which a threat is generated |
| `QSGW_ANOMALY_CRITICAL_THRESHOLD` | 0.85 | Score above which severity is CRITICAL  |

A connection with an anomaly score above the threshold generates a threat event. The severity is determined by the score magnitude and the specific threat type detected.

---

## Bot Detection

The bot detector uses behavioral analysis to distinguish automated traffic from legitimate human users. It operates independently of the anomaly detector and focuses on application-layer signals.

### Input Features

| Feature                 | Type   | Description                                       |
|-------------------------|--------|---------------------------------------------------|
| `user_agent`            | string | HTTP User-Agent header value                      |
| `requests_per_minute`   | float  | Request rate from the source IP                   |
| `unique_paths`          | int    | Number of distinct URL paths accessed             |
| `error_rate`            | float  | Fraction of requests resulting in errors (0--1)   |
| `avg_response_time_ms`  | float  | Average time the client takes to issue next request |

### Detection Signals

#### User-Agent Fingerprinting

The detector maintains a database of known bot user-agent patterns and applies heuristic analysis to unknown strings.

**High-confidence bot indicators:**

- Known scripting libraries: `python-requests`, `Go-http-client`, `curl`, `wget`, `scrapy`
- Known bot frameworks: `Googlebot`, `Bingbot`, `Baiduspider` (when not from expected IP ranges)
- Empty or missing User-Agent header
- User-Agent strings containing version numbers inconsistent with real browser releases
- Rapid User-Agent rotation from the same IP

#### Rate Analysis

Human users typically generate 1--10 requests per minute during active browsing. Rates significantly above this range indicate automation.

| Classification | Requests per Minute | Confidence |
|---------------|---------------------|------------|
| Normal human  | 1--30               | Low bot    |
| Elevated      | 30--100             | Medium bot |
| Automated     | 100--500            | High bot   |
| High-speed    | 500+                | Very high  |

#### Path Diversity Scoring

Automated crawlers typically access a much wider range of URL paths than human users. The detector evaluates path diversity relative to session duration.

**Indicators:**

- More than 50 unique paths in a 5-minute window
- Systematic path enumeration patterns (e.g., `/page/1`, `/page/2`, `/page/3`)
- Access to paths that are not linked from any page (directory scanning)

#### Error Rate Correlation

Bots often generate higher error rates than humans because they request paths that do not exist or lack proper authentication.

**Indicators:**

- Error rate above 0.15 combined with high request rate
- Concentrated 404 errors (path scanning)
- Concentrated 401/403 errors (credential probing)

### Composite Bot Score

The bot detector produces a composite score between 0.0 (human) and 1.0 (bot):

| Factor           | Weight | Description                              |
|------------------|--------|------------------------------------------|
| User-agent       | 0.30   | Known bot patterns and fingerprinting    |
| Request rate     | 0.35   | Requests per minute relative to baseline |
| Path diversity   | 0.25   | Unique path count relative to session    |
| Response timing  | 0.10   | Below-human response processing time     |

**Threshold configuration:**

| Parameter                  | Default | Description                          |
|----------------------------|---------|--------------------------------------|
| `QSGW_BOT_THRESHOLD`      | 0.70    | Score above which traffic is flagged |
| `QSGW_BOT_BLOCK_THRESHOLD`| 0.90    | Score above which traffic is blocked |

---

## Threat Types Reference

### QUANTUM_DOWNGRADE

| Property    | Value                                                                  |
|-------------|------------------------------------------------------------------------|
| Severity    | CRITICAL                                                               |
| Source      | Anomaly detector                                                       |
| Description | Client attempts to force negotiation of classical-only cipher suites, bypassing post-quantum protections |
| Indicators  | Classical cipher suite negotiation, TLS < 1.3, absence of ML-KEM     |
| Impact      | Exposes session to harvest-now-decrypt-later attacks                  |
| Response    | Connection rejection (PQC_ONLY), threat event generation, alerting    |

### WEAK_CIPHER

| Property    | Value                                                                  |
|-------------|------------------------------------------------------------------------|
| Severity    | HIGH                                                                   |
| Source      | Anomaly detector                                                       |
| Description | Client negotiates or requests deprecated or broken cipher suites       |
| Indicators  | RC4, DES, 3DES, MD5-based MACs, SHA1-based signatures                 |
| Impact      | Session vulnerable to known cryptographic attacks                      |
| Response    | Connection rejection, threat event generation                          |

### BOT_ATTACK

| Property    | Value                                                                  |
|-------------|------------------------------------------------------------------------|
| Severity    | HIGH                                                                   |
| Source      | Bot detector                                                           |
| Description | Automated traffic identified as data harvesting, scraping, or probing  |
| Indicators  | Scripting library UA, high RPM, high path diversity, low response time|
| Impact      | Resource exhaustion, data exfiltration, reconnaissance                 |
| Response    | Rate limiting, blocking, threat event generation                       |

### ANOMALOUS_TRAFFIC

| Property    | Value                                                                  |
|-------------|------------------------------------------------------------------------|
| Severity    | MEDIUM                                                                 |
| Source      | Anomaly detector                                                       |
| Description | Traffic patterns that deviate significantly from established baselines  |
| Indicators  | Rate spikes, unusual error rates, atypical access patterns             |
| Impact      | Potential precursor to targeted attacks                                |
| Response    | Monitoring, alerting, optional rate limiting                           |

### CERTIFICATE_ISSUE

| Property    | Value                                                                  |
|-------------|------------------------------------------------------------------------|
| Severity    | MEDIUM                                                                 |
| Source      | TLS engine                                                             |
| Description | TLS certificate validation failures on upstream or client connections  |
| Indicators  | Expired certificates, hostname mismatch, untrusted CA, revoked certs  |
| Impact      | Man-in-the-middle vulnerability, service disruption                   |
| Response    | Connection rejection for critical issues, alerting                     |

### REPLAY_ATTACK

| Property    | Value                                                                  |
|-------------|------------------------------------------------------------------------|
| Severity    | HIGH                                                                   |
| Source      | TLS engine                                                             |
| Description | Replayed TLS session tickets or handshake messages                     |
| Indicators  | Duplicate session IDs, reused nonces, identical handshake transcripts  |
| Impact      | Session hijacking, authentication bypass                               |
| Response    | Session invalidation, connection rejection, threat event generation    |

---

## Severity Levels

QSGW uses five severity levels to classify threat events:

| Level    | Priority | Description                                                | Action Required    |
|----------|----------|------------------------------------------------------------|--------------------|
| CRITICAL | 1        | Immediate threat to quantum safety or system integrity     | Immediate response |
| HIGH     | 2        | Active exploitation or significant vulnerability           | Prompt response    |
| MEDIUM   | 3        | Suspicious activity that may indicate an emerging threat   | Investigation      |
| LOW      | 4        | Minor policy violations or informational anomalies         | Monitoring         |
| INFO     | 5        | Normal operational events logged for audit purposes        | None               |

Severity levels are assigned automatically by the detection models based on the composite threat score and the specific threat type. `QUANTUM_DOWNGRADE` events are always `CRITICAL` regardless of score.

---

## Threat Mitigation

When a threat event is detected, it is persisted in the `threat_events` table and made available through the REST API and admin dashboard. Operators can mitigate threats through the API.

### Mitigation Workflow

```
1. Threat detected by AI engine
   |
   v
2. Threat event created (mitigated=false)
   |
   v
3. Operator reviews threat in dashboard or via API
   |
   v
4. Operator calls POST /api/v1/threats/{id}/mitigate
   |
   v
5. Threat marked as mitigated (mitigated=true, mitigated_at=timestamp)
   |
   v
6. Countermeasures applied (IP block, rate limit adjustment, etc.)
```

### API Mitigation

```bash
# List unmitigated threats
curl "http://localhost:8085/api/v1/threats?gateway_id=<id>" \
  -H "Authorization: Bearer <token>"

# Review a specific threat
curl http://localhost:8085/api/v1/threats/<threat_id> \
  -H "Authorization: Bearer <token>"

# Mitigate the threat
curl -X POST http://localhost:8085/api/v1/threats/<threat_id>/mitigate \
  -H "Authorization: Bearer <token>"
```

### Automatic Mitigation

Certain threat types trigger automatic countermeasures at the gateway level:

| Threat Type       | Automatic Action                                           |
|-------------------|------------------------------------------------------------|
| QUANTUM_DOWNGRADE | Connection rejected if policy is PQC_ONLY                  |
| WEAK_CIPHER       | Connection rejected; weak ciphers removed from offer list  |
| BOT_ATTACK        | Rate limiting applied; blocking if score > 0.90            |
| REPLAY_ATTACK     | Session invalidated; connection rejected                   |

Threats that are automatically mitigated at the connection level are still recorded as threat events for audit purposes.

---

## Integration with the Gateway

The AI engine is tightly integrated with the Rust gateway through an asynchronous analysis pipeline.

### Real-Time Analysis Flow

1. **TLS Handshake Completes:** The gateway records the negotiated cipher suite, TLS version, and handshake duration.

2. **Session Telemetry:** As the connection progresses, the gateway tracks requests per minute, error rates, and unique paths per source IP.

3. **Analysis Request:** Telemetry is sent to the AI engine via HTTP POST to `/api/v1/analyze-traffic` and `/api/v1/detect-bot`.

4. **Threat Assessment:** The AI engine returns threat scores and classifications.

5. **Event Recording:** If threats are detected, the control plane persists them in PostgreSQL.

6. **Enforcement:** The gateway applies any automatic countermeasures based on the threat type and severity.

### Latency Considerations

Analysis requests to the AI engine are sent asynchronously. The proxy path is not blocked by threat analysis. This means:

- Proxied request latency is not affected by AI engine processing time
- Threat detection results are available within 10--50ms of the triggering event
- Automatic countermeasures (e.g., IP blocking) are applied on subsequent requests, not the triggering request

### Failure Mode

If the AI engine is unavailable, the gateway continues to proxy requests without threat analysis. A `CERTIFICATE_ISSUE` event with severity `MEDIUM` is generated for the AI engine connection failure, and the admin dashboard shows the AI engine as unhealthy.

---

## Extending Detection

The AI engine is designed to be extensible. Custom detection modules can be added alongside the built-in anomaly and bot detectors.

### Adding a Custom Detector

1. **Create a new detector module** in the `ai-engine/` directory:

```
ai-engine/
  detectors/
    anomaly_detector.py
    bot_detector.py
    custom_detector.py     <-- new file
```

2. **Implement the detector interface:**

```python
from dataclasses import dataclass

@dataclass
class DetectionResult:
    is_threat: bool
    threat_type: str
    severity: str
    confidence: float
    description: str

class CustomDetector:
    def __init__(self):
        # Initialize model, thresholds, etc.
        pass

    def analyze(self, features: dict) -> DetectionResult:
        # Implement detection logic
        score = self._compute_score(features)
        return DetectionResult(
            is_threat=score > 0.7,
            threat_type="CUSTOM_THREAT",
            severity="MEDIUM",
            confidence=score,
            description="Custom threat detected"
        )

    def _compute_score(self, features: dict) -> float:
        # Custom scoring logic
        return 0.0
```

3. **Register the detector** in the FastAPI application by adding it to the analysis pipeline.

4. **Add the new threat type** to the control plane's threat type enum and database schema.

### Model Training

The anomaly and bot detectors use scikit-learn models that can be retrained on site-specific traffic data:

- Export traffic telemetry from the `tls_sessions` table
- Label samples as normal or anomalous
- Retrain using the provided training scripts in `ai-engine/training/`
- Replace the model files and restart the AI engine

---

## Alert Configuration

QSGW supports multiple alert delivery mechanisms for threat events.

### Structured Logging

All threat events are logged using structured JSON logging (Go zap logger) for ingestion by log aggregation systems such as ELK, Splunk, or Datadog.

```json
{
  "level": "warn",
  "ts": "2026-02-20T11:00:00.000Z",
  "caller": "threats/service.go:142",
  "msg": "threat detected",
  "threat_id": "thr_01HQ3PA10R6S8T0U2V4W5XYZ",
  "threat_type": "QUANTUM_DOWNGRADE",
  "severity": "CRITICAL",
  "source_ip": "203.0.113.42",
  "gateway_id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW"
}
```

### Webhook Integration

Configure a webhook URL to receive real-time threat notifications:

| Environment Variable       | Default | Description                            |
|----------------------------|---------|----------------------------------------|
| `QSGW_WEBHOOK_URL`        | --      | HTTPS endpoint for threat webhooks     |
| `QSGW_WEBHOOK_SECRET`     | --      | HMAC-SHA256 secret for webhook signing |
| `QSGW_WEBHOOK_MIN_SEVERITY`| `MEDIUM`| Minimum severity to trigger webhook   |

**Webhook payload format:**

```json
{
  "event": "threat.detected",
  "timestamp": "2026-02-20T11:00:00Z",
  "threat": {
    "id": "thr_01HQ3PA10R6S8T0U2V4W5XYZ",
    "type": "QUANTUM_DOWNGRADE",
    "severity": "CRITICAL",
    "source_ip": "203.0.113.42",
    "gateway_id": "gw_01HQ3K5V8N2M4P6R7S9T0UVW",
    "description": "Classical cipher suite negotiation detected"
  }
}
```

The webhook request includes an `X-QSGW-Signature` header containing an HMAC-SHA256 signature of the payload body, computed with the configured webhook secret.

### Admin Dashboard

The React admin dashboard (port 3003) provides a real-time threat monitoring view with:

- Live threat event feed with severity color coding
- Threat timeline visualization
- Per-gateway threat summary
- One-click mitigation actions
- Threat detail drill-down with full session metadata

### Integration with External SIEM

For enterprise deployments, QSGW threat events can be forwarded to Security Information and Event Management (SIEM) systems through:

1. **Log forwarding:** Ship structured JSON logs to your SIEM via Fluentd, Filebeat, or similar agents.
2. **Webhook bridge:** Point the webhook URL at a SIEM ingestion endpoint.
3. **Database polling:** Query the `threat_events` table directly from your SIEM's database connector.
