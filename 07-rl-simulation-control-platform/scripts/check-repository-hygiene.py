from __future__ import annotations

import re
from pathlib import Path
from urllib.parse import unquote

ROOT = Path(__file__).resolve().parents[1]
IGNORED_PARTS = {
    ".git",
    ".nuxt",
    ".output",
    ".pytest_cache",
    ".ruff_cache",
    ".tmp",
    ".venv",
    "__pycache__",
    "node_modules",
}
TEXT_SUFFIXES = {
    ".css",
    ".go",
    ".html",
    ".json",
    ".md",
    ".mjs",
    ".py",
    ".sql",
    ".toml",
    ".ts",
    ".vue",
    ".yaml",
    ".yml",
}
TEXT_NAMES = {".dockerignore", ".env.example", ".gitignore", "Dockerfile"}
BROKEN_TEXT = re.compile("\ufffd|\u00c3.|\u00c2.|\u00e2[\u0080-\u00bf]{2}")
SECRET_PATTERNS = (
    re.compile(r"postgres(?:ql)?://[^\s:]+:[^\s@]+@", re.IGNORECASE),
    re.compile(r"AWS_SECRET_ACCESS_KEY\s*=", re.IGNORECASE),
    re.compile(r"-----BEGIN (?:RSA |EC )?PRIVATE KEY-----"),
)
MARKDOWN_LINK = re.compile(r"!?\[[^\]]*\]\(([^)]+)\)")


def is_repository_text(path: Path) -> bool:
    relative = path.relative_to(ROOT)
    return not any(part in IGNORED_PARTS for part in relative.parts) and (
        path.suffix.lower() in TEXT_SUFFIXES or path.name in TEXT_NAMES
    )


def main() -> None:
    failures: list[str] = []
    inspected = 0
    for path in sorted(item for item in ROOT.rglob("*") if item.is_file()):
        if not is_repository_text(path):
            continue
        inspected += 1
        relative = path.relative_to(ROOT).as_posix()
        try:
            raw = path.read_bytes()
            text = raw.decode("utf-8")
        except UnicodeDecodeError:
            failures.append(f"{relative}: not valid UTF-8")
            continue
        if raw and not raw.endswith(b"\n"):
            failures.append(f"{relative}: missing final newline")
        for line_number, line in enumerate(text.splitlines(), start=1):
            if line.rstrip(" \t") != line:
                failures.append(f"{relative}:{line_number}: trailing whitespace")
        if BROKEN_TEXT.search(text):
            failures.append(f"{relative}: possible broken text encoding")
        if path.suffix.lower() == ".md":
            for target in MARKDOWN_LINK.findall(text):
                target = target.strip().split(" ", maxsplit=1)[0].strip("<>")
                if not target or target.startswith(("#", "http://", "https://", "mailto:")):
                    continue
                local_target = unquote(target.split("#", maxsplit=1)[0])
                if local_target and not (path.parent / local_target).resolve().exists():
                    failures.append(f"{relative}: broken local link {target}")
        if path.name != ".env.example":
            for pattern in SECRET_PATTERNS:
                match = pattern.search(text)
                if match and "${" not in match.group(0):
                    failures.append(f"{relative}: possible committed secret")

    if failures:
        raise SystemExit("\n".join(failures))
    print(f"OK - repository hygiene passed for {inspected} UTF-8 text files")


if __name__ == "__main__":
    main()
