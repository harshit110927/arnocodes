import re

CHATBOT_FALLBACK = "Insufficient context to answer precisely."

REQUIRED_CODE_HELPER_SECTIONS = [
    "Problem Understanding",
    "Intuition",
    "Brute Force Approach",
    "Optimization Path",
    "Final Optimal Solution",
    "Interview Simulation",
]


def enforce_chatbot_response(response: str) -> str:
    """Apply strict concise mode constraints for chatbot mode."""
    cleaned = " ".join((response or "").strip().split())
    if not cleaned:
        return CHATBOT_FALLBACK

    # Eliminate follow-up style outputs in concise mode.
    if "?" in cleaned:
        cleaned = cleaned.replace("?", ".")

    # Hard cap length to keep responses short/cost-safe in concise mode.
    max_chars = 320
    if len(cleaned) > max_chars:
        cleaned = cleaned[: max_chars - 1].rstrip() + "…"

    return cleaned or CHATBOT_FALLBACK


def enforce_code_helper_structure(response: str) -> str:
    """Guarantee all required sections exist and are emitted in strict order."""
    text = (response or "").strip()
    if not text:
        return _fallback_structured_response()

    parsed_sections = _extract_sections(text)
    normalized_parts = []
    for section in REQUIRED_CODE_HELPER_SECTIONS:
        content = parsed_sections.get(section, "Content unavailable.").strip()
        if not content:
            content = "Content unavailable."
        normalized_parts.append(f"## {section}\n\n{content}")

    return "\n\n".join(normalized_parts)


def _extract_sections(text: str) -> dict[str, str]:
    section_pattern = re.compile(
        r"(?:^|\n)\s{0,3}(?:##?\s*)?"
        r"(Problem Understanding|Intuition|Brute Force Approach|Optimization Path|Final Optimal Solution|Interview Simulation)\s*\n",
        re.IGNORECASE,
    )

    matches = list(section_pattern.finditer(text + "\n"))
    if not matches:
        return {}

    extracted: dict[str, str] = {}
    normalized_names = {s.lower(): s for s in REQUIRED_CODE_HELPER_SECTIONS}

    for index, match in enumerate(matches):
        raw_name = match.group(1).strip().lower()
        canonical_name = normalized_names.get(raw_name)
        if not canonical_name:
            continue
        start = match.end()
        end = matches[index + 1].start() if index + 1 < len(matches) else len(text)
        extracted[canonical_name] = text[start:end].strip()

    return extracted


def _fallback_structured_response() -> str:
    return "\n\n".join([f"## {section}\n\nContent unavailable." for section in REQUIRED_CODE_HELPER_SECTIONS])
