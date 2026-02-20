"""FastAPI application for the QSGW AI Engine."""

from __future__ import annotations

from fastapi import FastAPI
from pydantic import BaseModel

from qsgw_ai.anomaly_detector import AnomalyDetector
from qsgw_ai.anomaly_detector.detector import TrafficSample
from qsgw_ai.bot_detector import BotDetector
from qsgw_ai.bot_detector.detector import RequestFingerprint

app = FastAPI(title="QSGW AI Engine", version="0.1.0")

_anomaly = AnomalyDetector()
_bot = BotDetector()


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok", "service": "qsgw-ai"}


# ---------- Anomaly Detection ----------


class TrafficSampleInput(BaseModel):
    source_ip: str
    cipher_suite: str
    tls_version: str
    request_rate_rpm: float
    handshake_duration_ms: float
    is_pqc: bool
    path: str = "/"


class AnomalyResponse(BaseModel):
    is_anomaly: bool
    anomaly_type: str | None
    severity: str
    confidence: float
    description: str


@app.post("/api/v1/analyze-traffic", response_model=AnomalyResponse)
def analyze_traffic(req: TrafficSampleInput) -> AnomalyResponse:
    sample = TrafficSample(
        source_ip=req.source_ip,
        cipher_suite=req.cipher_suite,
        tls_version=req.tls_version,
        request_rate_rpm=req.request_rate_rpm,
        handshake_duration_ms=req.handshake_duration_ms,
        is_pqc=req.is_pqc,
        path=req.path,
    )
    result = _anomaly.analyze(sample)
    return AnomalyResponse(
        is_anomaly=result.is_anomaly,
        anomaly_type=result.anomaly_type.value if result.anomaly_type else None,
        severity=result.severity,
        confidence=result.confidence,
        description=result.description,
    )


# ---------- Bot Detection ----------


class BotFingerprintInput(BaseModel):
    source_ip: str
    user_agent: str
    path: str
    request_rate_rpm: float
    unique_paths_per_minute: int
    avg_response_time_ms: float
    has_valid_session: bool


class BotResponse(BaseModel):
    is_bot: bool
    confidence: float
    bot_category: str
    description: str


@app.post("/api/v1/detect-bot", response_model=BotResponse)
def detect_bot(req: BotFingerprintInput) -> BotResponse:
    fp = RequestFingerprint(
        source_ip=req.source_ip,
        user_agent=req.user_agent,
        path=req.path,
        request_rate_rpm=req.request_rate_rpm,
        unique_paths_per_minute=req.unique_paths_per_minute,
        avg_response_time_ms=req.avg_response_time_ms,
        has_valid_session=req.has_valid_session,
    )
    result = _bot.detect(fp)
    return BotResponse(
        is_bot=result.is_bot,
        confidence=result.confidence,
        bot_category=result.bot_category,
        description=result.description,
    )
