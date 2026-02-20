from qsgw_ai.anomaly_detector import AnomalyDetector
from qsgw_ai.anomaly_detector.detector import TrafficSample


def test_normal_traffic():
    detector = AnomalyDetector()
    sample = TrafficSample(
        source_ip="10.0.0.1",
        cipher_suite="TLS_ML-KEM-768_AES_256_GCM",
        tls_version="1.3",
        request_rate_rpm=100.0,
        handshake_duration_ms=50.0,
        is_pqc=True,
    )
    result = detector.analyze(sample)
    assert not result.is_anomaly
    assert result.severity == "LOW"


def test_quantum_downgrade():
    detector = AnomalyDetector()
    sample = TrafficSample(
        source_ip="10.0.0.1",
        cipher_suite="TLS_RSA_WITH_RC4_128_SHA",
        tls_version="1.0",
        request_rate_rpm=10.0,
        handshake_duration_ms=50.0,
        is_pqc=False,
    )
    result = detector.analyze(sample)
    assert result.is_anomaly
    assert result.anomaly_type.value == "QUANTUM_DOWNGRADE"
    assert result.severity == "CRITICAL"


def test_traffic_spike():
    detector = AnomalyDetector(max_request_rate=100.0)
    sample = TrafficSample(
        source_ip="10.0.0.1",
        cipher_suite="TLS_AES_256_GCM_SHA384",
        tls_version="1.3",
        request_rate_rpm=500.0,
        handshake_duration_ms=50.0,
        is_pqc=False,
    )
    result = detector.analyze(sample)
    assert result.is_anomaly
    assert result.anomaly_type.value == "TRAFFIC_SPIKE"


def test_abnormal_handshake():
    detector = AnomalyDetector(max_handshake_ms=1000.0)
    sample = TrafficSample(
        source_ip="10.0.0.1",
        cipher_suite="TLS_AES_256_GCM_SHA384",
        tls_version="1.3",
        request_rate_rpm=50.0,
        handshake_duration_ms=5000.0,
        is_pqc=False,
    )
    result = detector.analyze(sample)
    assert result.is_anomaly
    assert result.anomaly_type.value == "ABNORMAL_HANDSHAKE"
