#!/bin/bash
# SessionEnd hook (see .claude/settings.local.json): stops the local Go backend when the Claude
# session closes, so its lifetime matches the conversation's. caffeinate exits on its own (it was
# started with -w on this process's pid) but is force-killed too as a safety net.
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT" || exit 0

GIT_DIR=$(git rev-parse --git-dir 2>/dev/null || true)
case "$GIT_DIR" in
	*/worktrees/*) exit 0 ;;
esac

PORT=53890

PORT_PID=$(lsof -ti tcp:$PORT 2>/dev/null || true)
[ -n "$PORT_PID" ] && kill $PORT_PID 2>/dev/null || true

pkill -f "[g]o run ./cmd/service --dev" 2>/dev/null || true
pkill -f "[m]ake build-ui" 2>/dev/null || true
pkill -f "[c]affeinate -d -u -w" 2>/dev/null || true

exit 0
