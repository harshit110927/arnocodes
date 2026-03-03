from dataclasses import dataclass
from datetime import timedelta

from config import AIConfig


@dataclass(frozen=True)
class ModeConfig:
    mode: str
    per_user_limit: int
    window: timedelta
    max_output_tokens: int


class ModeHandler:
    @staticmethod
    def get(mode: str) -> ModeConfig:
        if mode == "chatbot":
            return ModeConfig(
                mode="chatbot",
                per_user_limit=AIConfig.CHATBOT_MAX_PER_HOUR,
                window=timedelta(hours=1),
                max_output_tokens=AIConfig.CHATBOT_MAX_OUTPUT_TOKENS,
            )
        if mode == "code_helper":
            return ModeConfig(
                mode="code_helper",
                per_user_limit=AIConfig.CODE_HELPER_MAX_PER_DAY,
                window=timedelta(days=1),
                max_output_tokens=AIConfig.CODE_HELPER_MAX_OUTPUT_TOKENS,
            )
        raise ValueError("Invalid mode")
