"""Traffic anomaly detection for the Quantum-Safe Gateway.

Detects unusual traffic patterns that may indicate quantum downgrade attacks,
cipher suite manipulation, or abnormal handshake behaviour.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum


class AnomalyType(str, Enum):
    QUANTUM_DOWNGRADE = "QUANTUM_DOWNGRADE"
    CIPHER_MANIPULATION = "CIPHER_MANIPULATION"
    ABNORMAL_HANDSHAKE = "ABNORMAL_HANDSHAKE"
    TRAFFIC_SPIKE = "TRAFFIC_SPIKE"
    UNUSUAL_SOURCE = "UNUSUAL_SOURCE"


@dataclass
class TrafficSample:
    source_ip: str
    cipher_suite: str
    tls_version: str
    request_rate_rpm: float
    handshake_duration_ms: float
    is_pqc: bool
    path: str = "/"


@dataclass
class AnomalyResult:
    is_anomaly: bool
    anomaly_type: AnomalyType | None
    severity: str  # CRITICAL, HIGH, MEDIUM, LOW
    confidence: float  # 0.0 - 1.0
    description: str


class AnomalyDetector:
    """Rule-based anomaly detector with configurable thresholds."""

    def __init__(
        self,
        max_request_rate: float = 600.0,
        max_handshake_ms: float = 5000.0,
        min_tls_version: str = "1.3",
    ):
        self.max_request_rate = max_request_rate
        self.max_handshake_ms = max_handshake_ms
        self.min_tls_version = min_tls_version

    def analyze(self, sample: TrafficSample) -> AnomalyResult:
        # Check for quantum downgrade attack
        if self._is_downgrade(sample):
            return AnomalyResult(
                is_anomaly=True,
                anomaly_type=AnomalyType.QUANTUM_DOWNGRADE,
                severity="CRITICAL",
                confidence=0.9,
                description=f"Potential quantum downgrade: {sample.cipher_suite} on TLS {sample.tls_version}",
            )

        # Check for traffic spike
        if sample.request_rate_rpm > self.max_request_rate:
            severity = "HIGH" if sample.request_rate_rpm > self.max_request_rate * 3 else "MEDIUM"
            return AnomalyResult(
                is_anomaly=True,
                anomaly_type=AnomalyType.TRAFFIC_SPIKE,
                severity=severity,
                confidence=0.85,
                description=f"Traffic spike: {sample.request_rate_rpm:.0f} req/min from {sample.source_ip}",
            )

        # Check for abnormal handshake
        if sample.handshake_duration_ms > self.max_handshake_ms:
            return AnomalyResult(
                is_anomaly=True,
                anomaly_type=AnomalyType.ABNORMAL_HANDSHAKE,
                severity="MEDIUM",
                confidence=0.7,
                description=f"Abnormal handshake duration: {sample.handshake_duration_ms:.0f}ms",
            )

        return AnomalyResult(
            is_anomaly=False,
            anomaly_type=None,
            severity="LOW",
            confidence=0.0,
            description="Normal traffic",
        )

    def _is_downgrade(self, sample: TrafficSample) -> bool:
        weak_ciphers = ["RC4", "DES", "3DES", "MD5", "SHA1"]
        if any(c in sample.cipher_suite for c in weak_ciphers):
            return True
        if sample.tls_version in ("1.0", "1.1"):
            return True
        return False
