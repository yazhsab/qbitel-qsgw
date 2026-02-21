# Security Policy

The security of QSGW is a top priority. This document describes our security practices, supported versions, vulnerability reporting process, and the security features built into the Quantum-Safe Gateway.

## Table of Contents

- [Supported Versions](#supported-versions)
- [Reporting Vulnerabilities](#reporting-vulnerabilities)
- [Response Timeline](#response-timeline)
- [PQC TLS Security Model](#pqc-tls-security-model)
- [Threat Detection Capabilities](#threat-detection-capabilities)
- [Security Features](#security-features)
- [Responsible Disclosure](#responsible-disclosure)

---

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 0.1.x   | Yes (current)     |

Security patches are provided for the latest minor release. We recommend always running the most recent version.

---

## Reporting Vulnerabilities

**Do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability in QSGW, please report it responsibly by emailing:

**[security@qbitel.dev](mailto:security@qbitel.dev)**

### What to Include

- A description of the vulnerability and its potential impact.
- Detailed steps to reproduce the issue.
- The affected component(s): gateway, control plane, AI engine, admin dashboard, crypto, or TLS.
- Your assessment of the severity (Critical, High, Medium, Low).
- Any suggested mitigation or fix.

### Encryption

If the vulnerability involves sensitive information, you may encrypt your report using our PGP key, which is available at [https://qbitel.dev/.well-known/pgp-key.txt](https://qbitel.dev/.well-known/pgp-key.txt).

---

## Response Timeline

We are committed to addressing security vulnerabilities promptly:

| Stage                        | Target Timeline     |
|------------------------------|---------------------|
| Acknowledgment of report     | Within 24 hours     |
| Initial assessment           | Within 48 hours     |
| Confirmed vulnerability fix  | Within 7 days       |
| Security advisory published  | Within 14 days      |
| Patch release                | Within 14 days      |

For critical vulnerabilities affecting cryptographic security or TLS integrity, we aim to provide a fix within 72 hours of confirmation.

---

## PQC TLS Security Model

QSGW implements a defense-in-depth approach to post-quantum TLS:

### Quantum Threat Model

QSGW is designed to protect against the following quantum threats:

- **Harvest-now, decrypt-later (HNDL):** Adversaries capturing encrypted traffic today for decryption by future quantum computers. QSGW mitigates this by using post-quantum key exchange (ML-KEM) that resists quantum attacks.
- **Quantum key recovery:** Future quantum computers using Shor's algorithm to break RSA and ECDH key exchange. QSGW uses lattice-based key encapsulation (ML-KEM) that is not vulnerable to Shor's algorithm.
- **Quantum signature forgery:** Future quantum computers forging classical digital signatures. QSGW supports lattice-based (ML-DSA) and hash-based (SLH-DSA) signatures that resist quantum attacks.

### Cryptographic Standards

All post-quantum algorithms implemented in QSGW are based on finalized NIST standards:

| Standard  | Algorithm | Purpose            |
|-----------|-----------|---------------------|
| FIPS 203  | ML-KEM    | Key encapsulation   |
| FIPS 204  | ML-DSA    | Digital signatures  |
| FIPS 205  | SLH-DSA   | Hash-based signatures |

### TLS Policy Enforcement

The gateway enforces TLS policies at the handshake level. Connections that do not meet the configured policy requirements are rejected before any application data is exchanged. Policy violations generate threat events for monitoring and audit.

---

## Threat Detection Capabilities

The AI-powered threat detection system provides continuous monitoring for:

- **Quantum downgrade attacks:** Detection of attempts to force classical-only cipher suites.
- **Weak cipher usage:** Identification of deprecated algorithms (RC4, DES, 3DES, MD5, SHA1).
- **Bot attacks:** Behavioral analysis to identify automated traffic.
- **Anomalous traffic patterns:** Rate spike detection and unusual access patterns.
- **Certificate issues:** Expired, invalid, or misconfigured TLS certificates.
- **Replay attacks:** Detection of replayed TLS sessions.

All detected threats are logged, persisted in the database, and available through the API and admin dashboard.

---

## Security Features

### FIPS 203/204/205 Compliant Cryptography

The `crypto/` crate implements post-quantum cryptographic algorithms conforming to the finalized NIST FIPS standards. Implementations are tested against known answer test (KAT) vectors from the NIST specifications.

### Constant-Time Authentication Comparison

All authentication credential comparisons (JWT signatures, API keys) use constant-time comparison functions to prevent timing side-channel attacks. The implementation avoids early-exit patterns that could leak information about the validity of partial inputs.

```
// Timing-safe comparison prevents attackers from
// determining the correct key one byte at a time.
```

### Memory Zeroization for Key Material

All cryptographic key material (private keys, shared secrets, session keys) is explicitly zeroized in memory when it is no longer needed. The `crypto/` and `tls/` crates use Rust's `zeroize` crate to ensure that sensitive data does not persist in memory after use.

This protects against:

- Memory dump attacks
- Core dump analysis
- Cold boot attacks

### Rate Limiting

Per-IP sliding-window rate limiting protects against brute-force attacks, credential stuffing, and denial-of-service attempts. Rate limits are configurable at both the global and per-route levels.

### Security Headers

The gateway injects the following security headers into every response:

| Header                        | Value                                         |
|-------------------------------|-----------------------------------------------|
| `Strict-Transport-Security`   | `max-age=63072000; includeSubDomains`         |
| `Content-Security-Policy`     | `default-src 'self'`                          |
| `X-Frame-Options`             | `DENY`                                        |
| `X-Content-Type-Options`      | `nosniff`                                     |
| `X-XSS-Protection`           | `0`                                            |
| `Referrer-Policy`             | `strict-origin-when-cross-origin`             |

These headers protect against clickjacking, MIME type sniffing, and cross-site scripting.

### Request Body Limits

The gateway enforces maximum request body sizes to prevent resource exhaustion attacks. The default limit is 10 MB, configurable via the `QSGW_MAX_BODY_SIZE` environment variable.

### Parameterized SQL

All database queries in the Go control plane use parameterized queries via pgx. No string concatenation or interpolation is used in SQL statements, preventing SQL injection attacks.

```go
// All queries use parameterized placeholders
row := pool.QueryRow(ctx, "SELECT * FROM gateways WHERE id = $1", id)
```

### Additional Security Measures

- **JWT expiration enforcement:** Tokens are validated for expiration on every request.
- **API key hashing:** API keys are stored as bcrypt hashes in the database.
- **Audit logging:** All administrative actions are recorded in the `qsgw_audit_log` table.
- **TLS session tracking:** All TLS sessions are recorded for forensic analysis.
- **Connection limits:** Maximum concurrent connections prevent resource exhaustion.
- **Upstream TLS verification:** TLS certificates for upstream services are verified by default.

---

## Responsible Disclosure

We ask that security researchers follow responsible disclosure practices:

1. **Report privately:** Send vulnerability reports to [security@qbitel.dev](mailto:security@qbitel.dev), not to public issue trackers.
2. **Allow time for a fix:** Give us a reasonable amount of time to address the vulnerability before public disclosure. We target 14 days for a patch release.
3. **Do not exploit:** Do not access, modify, or delete data belonging to other users while researching vulnerabilities.
4. **Minimize impact:** Make a good-faith effort to avoid disrupting production systems during testing.

### Recognition

We appreciate the efforts of security researchers who help keep QSGW safe. With your permission, we will acknowledge your contribution in the security advisory and release notes. We do not currently operate a formal bug bounty program.

### Safe Harbor

We will not pursue legal action against security researchers who:

- Act in good faith and follow this responsible disclosure policy.
- Avoid privacy violations, data destruction, and disruption of services.
- Report vulnerabilities promptly and provide sufficient detail for reproduction.

---

## Contact

For security-related inquiries:

- **Email:** [security@qbitel.dev](mailto:security@qbitel.dev)
- **PGP Key:** [https://qbitel.dev/.well-known/pgp-key.txt](https://qbitel.dev/.well-known/pgp-key.txt)
