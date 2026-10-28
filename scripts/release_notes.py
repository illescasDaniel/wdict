#!/usr/bin/env python3
"""Print the changelog section for a release version."""

import re
import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 2 or not re.fullmatch(r"(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)", sys.argv[1]):
        print("Usage: release_notes.py X.Y.Z", file=sys.stderr)
        return 1
    changelog = Path(__file__).resolve().parent.parent / "CHANGELOG.md"
    text = changelog.read_text(encoding="utf-8")
    header = re.compile(r"^## \[([^\]]+)\] - \d{4}-\d{2}-\d{2}\s*$", re.MULTILINE)
    sections = list(header.finditer(text))
    matches = [i for i, item in enumerate(sections) if item.group(1) == sys.argv[1]]
    if len(matches) != 1:
        print(f"Expected exactly one changelog section for {sys.argv[1]}", file=sys.stderr)
        return 1
    index = matches[0]
    end = sections[index + 1].start() if index + 1 < len(sections) else len(text)
    notes = text[sections[index].end():end].strip()
    if not notes:
        print("Release notes must not be empty", file=sys.stderr)
        return 1
    print(notes)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
