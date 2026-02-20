"""Bot and automated traffic detection for the Quantum-Safe Gateway.

Identifies automated traffic that may be attempting to harvest encrypted data
(Harvest Now, Decrypt Later) or probe for quantum-vulnerable endpoints.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass
class RequestFingerprint:
    source_ip: str
    user_agent: str
    path: str
    request_rate_rpm: float
    unique_paths_per_minute: int
    avg_response_time_ms: float
    has_valid_session: bool


@dataclass
class BotDetectionResult:
    is_bot: bool
    confidence: float  # 0.0 - 1.0
    bot_category: str  # CRAWLER, SCRAPER, ATTACKER, UNKNOWN, LEGITIMATE
    description: str


class BotDetector:
    """Heuristic-based bot detector."""

    def __init__(
        self,
        rate_threshold: float = 300.0,
        path_diversity_threshold: int = 50,
    ):
        self.rate_threshold = rate_threshold
        self.path_diversity_threshold = path_diversity_threshold
        self._known_bot_agents = [
            "bot", "crawler", "spider", "scraper", "curl", "wget", "python-requests",
            "go-http-client", "java/", "libwww",
        ]

    def detect(self, fingerprint: RequestFingerprint) -> BotDetectionResult:
        # Check user agent against known bots
        ua_lower = fingerprint.user_agent.lower()
        for bot_pattern in self._known_bot_agents:
            if bot_pattern in ua_lower:
                return BotDetectionResult(
                    is_bot=True,
                    confidence=0.95,
                    bot_category="CRAWLER",
                    description=f"Known bot user-agent: {fingerprint.user_agent}",
                )

        # Empty or suspicious user agent
        if not fingerprint.user_agent or len(fingerprint.user_agent) < 10:
            return BotDetectionResult(
                is_bot=True,
                confidence=0.8,
                bot_category="UNKNOWN",
                description="Missing or suspiciously short user-agent",
            )

        # High request rate
        if fingerprint.request_rate_rpm > self.rate_threshold:
            category = "ATTACKER" if fingerprint.request_rate_rpm > self.rate_threshold * 5 else "SCRAPER"
            return BotDetectionResult(
                is_bot=True,
                confidence=0.85,
                bot_category=category,
                description=f"High request rate: {fingerprint.request_rate_rpm:.0f} req/min",
            )

        # High path diversity (scanning behaviour)
        if fingerprint.unique_paths_per_minute > self.path_diversity_threshold:
            return BotDetectionResult(
                is_bot=True,
                confidence=0.75,
                bot_category="SCRAPER",
                description=f"High path diversity: {fingerprint.unique_paths_per_minute} unique paths/min",
            )

        return BotDetectionResult(
            is_bot=False,
            confidence=0.0,
            bot_category="LEGITIMATE",
            description="Traffic appears legitimate",
        )
