#!/bin/bash
# Personal dev-loop hook (see .claude/settings.local.json Stop hook): rebuilds the UI and
# restarts the Go backend after a turn touches source, so the local instance stays on what was
# just built without restarting it by hand. Never prompts — only session-start asks before
# killing an orphan.
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT" || exit 0

GIT_DIR=$(git rev-parse --git-dir 2>/dev/null || true)
case "$GIT_DIR" in
	*/worktrees/*) exit 0 ;;
esac

PORT=53890
MARKER=".claude/hooks/.last-serve-restart"
LOG=".claude/hooks/serve.log"

if [ -f "$MARKER" ]; then
	CHANGED=$(find . \( -name node_modules -o -name .git -o -name graphify-out -o -name worktrees -o -name dist \) -prune -o \
		\( -name '*.go' -o -name '*.ts' -o -name '*.tsx' -o -name '*.css' -o -name '*.sql' -o -name '*.proto' \) \
		-newer "$MARKER" -print 2>/dev/null | head -n 1)
else
	CHANGED="first-run"
fi

if [ -z "$CHANGED" ]; then
	exit 0
fi

PORT_PID=$(lsof -ti tcp:$PORT 2>/dev/null || true)
[ -n "$PORT_PID" ] && kill $PORT_PID 2>/dev/null || true

touch "$MARKER"

nohup bash -c "make build-ui && go run ./cmd/service --dev" >"$LOG" 2>&1 </dev/null &
SERVE_PID=$!
disown

caffeinate -d -u -w "$SERVE_PID" >/dev/null 2>&1 &
disown

exit 0
