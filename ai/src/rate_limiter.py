from collections import defaultdict, deque
from datetime import datetime, timedelta, timezone
from threading import Lock


class SlidingWindowRateLimiter:
    """In-memory sliding-window rate limiter for per-user and global limits."""

    def __init__(self):
        self._lock = Lock()
        self._user_mode_requests = defaultdict(deque)
        self._global_requests = deque()

    def _trim(self, queue: deque, window: timedelta, now: datetime) -> None:
        threshold = now - window
        while queue and queue[0] < threshold:
            queue.popleft()

    def allow(self, user_id: str, mode: str, mode_limit: int, mode_window: timedelta, global_daily_limit: int):
        now = datetime.now(timezone.utc)
        global_window = timedelta(days=1)

        with self._lock:
            self._trim(self._global_requests, global_window, now)
            if len(self._global_requests) >= global_daily_limit:
                return False, "Global daily AI request limit reached.", "global"

            key = f"{user_id}:{mode}"
            mode_queue = self._user_mode_requests[key]
            self._trim(mode_queue, mode_window, now)
            if len(mode_queue) >= mode_limit:
                return False, f"Rate limit exceeded for {mode} mode.", "per_user"

            mode_queue.append(now)
            self._global_requests.append(now)
            return True, None, None
