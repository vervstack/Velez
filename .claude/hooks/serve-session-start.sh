#!/bin/bash
# SessionStart hook (see .claude/settings.local.json): builds the UI, embeds it into the Go
# binary's dist dir, and brings up the Go backend for the session.
#
#   - worktree checkout          -> no-op, never touches serve here
#   - port 53890 free            -> build + start, silently
#   - port busy, source=startup  -> emit additionalContext telling Claude to ask the user
#                                    whether to keep the running instance or restart fresh
#   - port busy, source=resume   -> leave it (no nagging on every resume/compact)
#   - invoked with --restart     -> kill whatever holds 53890 and start fresh
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

start_serve() {
	touch "$MARKER"
	nohup bash -c "make build-ui && go run ./cmd/service --dev" >"$LOG" 2>&1 </dev/null &
	local pid=$!
	disown
	caffeinate -d -u -w "$pid" >/dev/null 2>&1 &
	disown
}

kill_serve() {
	local port_pid
	port_pid=$(lsof -ti tcp:$PORT 2>/dev/null || true)
	[ -n "$port_pid" ] && kill $port_pid 2>/dev/null || true
	pkill -f "[g]o run ./cmd/service --dev" 2>/dev/null || true
	pkill -f "[m]ake build-ui" 2>/dev/null || true
	local i
	for i in $(seq 1 20); do
		lsof -ti tcp:$PORT >/dev/null 2>&1 || return 0
		sleep 0.25
	done
}

if [ "${1:-}" = "--restart" ]; then
	kill_serve
	start_serve
	exit 0
fi

INPUT=$(cat 2>/dev/null || true)
SOURCE=$(printf '%s' "$INPUT" | jq -r '.source // "startup"' 2>/dev/null || echo startup)

PORT_PID=$(lsof -ti tcp:$PORT 2>/dev/null | head -1 || true)

if [ -z "$PORT_PID" ]; then
	start_serve
	exit 0
fi

case "$SOURCE" in
	startup | clear)
		UPTIME=$(ps -o etime= -p "$PORT_PID" 2>/dev/null | tr -d ' ')
		CTX="Velez is already listening on :${PORT} (PID ${PORT_PID}, up ${UPTIME:-unknown}) — likely an orphan from a previous session. Before anything else this session, call AskUserQuestion asking whether to keep this running instance or restart it fresh. If they pick restart, run: bash .claude/hooks/serve-session-start.sh --restart"
		jq -cn --arg c "$CTX" '{hookSpecificOutput:{hookEventName:"SessionStart",additionalContext:$c}}'
		;;
esac

exit 0
