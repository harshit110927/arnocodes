import os


class AIConfig:
    """Configuration surface for AI module."""

    GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
    GEMINI_MODEL = os.getenv("GEMINI_MODEL", "gemini-2.5-flash")

    CHATBOT_MAX_PER_HOUR = int(os.getenv("AI_CHATBOT_MAX_PER_HOUR", "3"))
    CODE_HELPER_MAX_PER_DAY = int(os.getenv("AI_CODE_HELPER_MAX_PER_DAY", "5"))
    GLOBAL_MAX_PER_DAY = int(os.getenv("AI_GLOBAL_MAX_PER_DAY", "300"))

    CHATBOT_MAX_OUTPUT_TOKENS = int(os.getenv("AI_CHATBOT_MAX_OUTPUT_TOKENS", "180"))
    CODE_HELPER_MAX_OUTPUT_TOKENS = int(os.getenv("AI_CODE_HELPER_MAX_OUTPUT_TOKENS", "900"))

    REQUEST_TIMEOUT_SECONDS = float(os.getenv("AI_REQUEST_TIMEOUT_SECONDS", "20"))
