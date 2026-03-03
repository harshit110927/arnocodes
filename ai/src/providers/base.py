from abc import ABC, abstractmethod


class LLMProvider(ABC):
    @abstractmethod
    def generate(self, prompt: str, config: dict) -> dict:
        raise NotImplementedError
