from config import AIConfig
from errors import InvalidInputError, RateLimitExceededError
from logging_wrapper import log_ai_event
from mode_handler import ModeHandler
from prompt_builder import build_prompt
from rate_limiter import SlidingWindowRateLimiter
from response_formatter import enforce_chatbot_response, enforce_code_helper_structure


class AIOrchestrator:
    """Coordinates validation, limiting, provider generation and response shaping."""

    def __init__(self, provider):
        self.provider = provider
        self.rate_limiter = SlidingWindowRateLimiter()

    def process(self, user_id: str, mode: str, text: str) -> dict:
        normalized_user_id = _normalize_user_id(user_id)
        normalized_text = _normalize_text(text)

        try:
            mode_cfg = ModeHandler.get(mode)
        except ValueError as exc:
            raise InvalidInputError("mode must be 'chatbot' or 'code_helper'") from exc

        allowed, reason, limit_type = self.rate_limiter.allow(
            user_id=normalized_user_id,
            mode=mode_cfg.mode,
            mode_limit=mode_cfg.per_user_limit,
            mode_window=mode_cfg.window,
            global_daily_limit=AIConfig.GLOBAL_MAX_PER_DAY,
        )
        if not allowed:
            log_ai_event(
                "rate_limit_hit",
                {
                    "userId": normalized_user_id,
                    "mode": mode,
                    "rate_limit_type": limit_type,
                },
            )
            raise RateLimitExceededError(reason)

        system_instruction, prompt = build_prompt(mode, normalized_text)
        result = self.provider.generate(
            prompt=prompt,
            config={
                "system_instruction": system_instruction,
                "max_output_tokens": mode_cfg.max_output_tokens,
                "temperature": 0.1 if mode == "chatbot" else 0.25,
            },
        )

        model_text = result.get("text", "")
        if mode == "chatbot":
            formatted = enforce_chatbot_response(model_text)
        else:
            formatted = enforce_code_helper_structure(model_text)

        input_tokens = int(result.get("input_tokens", _estimate_tokens(prompt)))
        output_tokens = int(result.get("output_tokens", _estimate_tokens(formatted)))

        log_ai_event(
            "ai_request",
            {
                "userId": normalized_user_id,
                "mode": mode,
                "input_token_estimate": input_tokens,
                "output_token_estimate": output_tokens,
            },
        )

        return {
            "mode": mode,
            "response": formatted,
            "usage": {
                "inputTokenEstimate": input_tokens,
                "outputTokenEstimate": output_tokens,
            },
        }


def _normalize_user_id(user_id: str) -> str:
    if not user_id or not isinstance(user_id, str):
        raise InvalidInputError("userId is required")
    normalized = user_id.strip()
    if not normalized:
        raise InvalidInputError("userId is required")
    if len(normalized) > 128:
        raise InvalidInputError("userId is too long")
    return normalized


def _normalize_text(text: str) -> str:
    if not text or not isinstance(text, str):
        raise InvalidInputError("text is required")
    normalized = text.strip()
    if not normalized:
        raise InvalidInputError("text is required")
    if len(normalized) > 12000:
        raise InvalidInputError("text exceeds allowed length")
    return normalized


def _estimate_tokens(text: str) -> int:
    return max(1, len((text or "").split()))
