import requests

from config import AIConfig
from errors import ProviderError, TimeoutError
from providers.base import LLMProvider


class GeminiProvider(LLMProvider):
    """Gemini API adapter implementing swappable LLMProvider interface."""

    def __init__(self):
        self.api_key = AIConfig.GEMINI_API_KEY
        self.model = AIConfig.GEMINI_MODEL
        self.timeout = AIConfig.REQUEST_TIMEOUT_SECONDS

    def generate(self, prompt: str, config: dict) -> dict:
        if not self.api_key:
            raise ProviderError("AI provider is not configured.")

        url = f"https://generativelanguage.googleapis.com/v1beta/models/{self.model}:generateContent"
        headers = {
            "Content-Type": "application/json",
            "x-goog-api-key": self.api_key,
        }
        payload = {
            "system_instruction": {"parts": [{"text": config.get("system_instruction", "")}]},
            "contents": [{"parts": [{"text": prompt}]}],
            "generationConfig": {
                "maxOutputTokens": config.get("max_output_tokens", 256),
                "temperature": config.get("temperature", 0.2),
            },
        }

        try:
            response = requests.post(url, headers=headers, json=payload, timeout=self.timeout)
        except requests.Timeout as exc:
            raise TimeoutError() from exc
        except requests.RequestException as exc:
            raise ProviderError() from exc

        if response.status_code >= 400:
            raise ProviderError()

        body = response.json()
        text = (
            body.get("candidates", [{}])[0]
            .get("content", {})
            .get("parts", [{}])[0]
            .get("text", "")
        )

        usage = body.get("usageMetadata", {})
        return {
            "text": text,
            "input_tokens": usage.get("promptTokenCount", _estimate_tokens(prompt)),
            "output_tokens": usage.get("candidatesTokenCount", _estimate_tokens(text)),
        }


def _estimate_tokens(text: str) -> int:
    return max(1, len((text or "").split()))
