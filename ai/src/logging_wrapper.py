import json
import logging
from datetime import datetime, timezone


logger = logging.getLogger("ai-service")
if not logger.handlers:
    handler = logging.StreamHandler()
    handler.setFormatter(logging.Formatter("%(message)s"))
    logger.addHandler(handler)
logger.setLevel(logging.INFO)


def log_ai_event(event: str, payload: dict) -> None:
    record = {
        "event": event,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        **payload,
    }
    logger.info(json.dumps(record))
