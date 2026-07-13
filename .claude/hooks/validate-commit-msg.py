#!/usr/bin/env python3
# PreToolUse hook: enforce the repo's `[Area] Category: description` commit
# message format on `git commit` Bash calls. Fails open (no block) whenever
# the message can't be confidently extracted from the shell command, e.g.
# `git commit --amend` with no `-m` (editor-composed messages aren't visible
# to a PreToolUse hook).
import json
import re
import sys

PATTERN = re.compile(
    r"^\[[^\[\]]+\] (bug fixed|refactor|docs|chore|perf|[A-Z][A-Za-z0-9]*( [A-Za-z0-9]+)*): .+$"
)

HELP = (
    "Commit message does not match the required format: [Area] Category: description\n\n"
    "  Area     - the part of the system changed, e.g. Tract, Connections, Auth\n"
    "  Category - one of: bug fixed, refactor, docs, chore, perf\n"
    "             or, for feature work, a Capitalized sub-area name (no literal 'feature')\n\n"
    "Examples:\n"
    "  [Tract] bug fixed: connections created on the caller side, not the receiver\n"
    "  [Tract] Canvas: added drag-select for nodes"
)


def extract_message(command):
    heredoc = re.search(r"<<[-~]?['\"]?(\w+)['\"]?\r?\n(.*?)\r?\n\1\b", command, re.S)
    if heredoc:
        return heredoc.group(2)

    dquoted = re.search(r"-m\s+\"((?:[^\"\\]|\\.)*)\"", command)
    if dquoted:
        return dquoted.group(1).replace('\\"', '"')

    squoted = re.search(r"-m\s+'([^']*)'", command)
    if squoted:
        return squoted.group(1)

    return None


def main():
    data = json.load(sys.stdin)
    command = data.get("tool_input", data).get("command", "")

    if "git commit" not in command:
        return

    message = extract_message(command)
    if message is None:
        return

    stripped = message.strip()
    first_line = stripped.splitlines()[0] if stripped else ""
    if not first_line or first_line.startswith("Merge "):
        return

    if not PATTERN.match(first_line):
        print(json.dumps({"decision": "block", "reason": HELP}))


if __name__ == "__main__":
    main()
