class AIServiceError(Exception):
    def __init__(self, code: str, message: str, status_code: int = 400):
        super().__init__(message)
        self.code = code
        self.message = message
        self.status_code = status_code


class InvalidInputError(AIServiceError):
    def __init__(self, message: str):
        super().__init__("INVALID_INPUT", message, 400)


class RateLimitExceededError(AIServiceError):
    def __init__(self, message: str):
        super().__init__("RATE_LIMIT_EXCEEDED", message, 429)


class ProviderError(AIServiceError):
    def __init__(self, message: str = "AI provider unavailable. Please try again later."):
        super().__init__("PROVIDER_ERROR", message, 502)


class TimeoutError(AIServiceError):
    def __init__(self, message: str = "AI request timed out. Please try again."):
        super().__init__("TIMEOUT", message, 504)
