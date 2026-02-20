from qsgw_ai.bot_detector import BotDetector
from qsgw_ai.bot_detector.detector import RequestFingerprint


def test_known_bot_agent():
    detector = BotDetector()
    fp = RequestFingerprint(
        source_ip="10.0.0.1",
        user_agent="Googlebot/2.1",
        path="/",
        request_rate_rpm=50.0,
        unique_paths_per_minute=10,
        avg_response_time_ms=100.0,
        has_valid_session=False,
    )
    result = detector.detect(fp)
    assert result.is_bot
    assert result.bot_category == "CRAWLER"


def test_legitimate_traffic():
    detector = BotDetector()
    fp = RequestFingerprint(
        source_ip="10.0.0.1",
        user_agent="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/120",
        path="/dashboard",
        request_rate_rpm=20.0,
        unique_paths_per_minute=5,
        avg_response_time_ms=200.0,
        has_valid_session=True,
    )
    result = detector.detect(fp)
    assert not result.is_bot
    assert result.bot_category == "LEGITIMATE"


def test_high_rate_attacker():
    detector = BotDetector(rate_threshold=100.0)
    fp = RequestFingerprint(
        source_ip="10.0.0.1",
        user_agent="Mozilla/5.0 (Windows NT 10.0) Chrome/120",
        path="/api/v1/keys",
        request_rate_rpm=600.0,
        unique_paths_per_minute=5,
        avg_response_time_ms=50.0,
        has_valid_session=False,
    )
    result = detector.detect(fp)
    assert result.is_bot
    assert result.bot_category == "ATTACKER"


def test_empty_user_agent():
    detector = BotDetector()
    fp = RequestFingerprint(
        source_ip="10.0.0.1",
        user_agent="",
        path="/",
        request_rate_rpm=10.0,
        unique_paths_per_minute=1,
        avg_response_time_ms=100.0,
        has_valid_session=False,
    )
    result = detector.detect(fp)
    assert result.is_bot
    assert result.bot_category == "UNKNOWN"
