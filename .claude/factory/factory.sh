#!/usr/bin/env bash
# factory.sh — pick a task and run it through the agent pipeline
# Usage:
#   ./factory.sh          → interactive task picker
#   ./factory.sh 003      → run specific task by ID
#   ./factory.sh list     → show all tasks and their status

set -euo pipefail

OLLAMA_HOST="${OLLAMA_HOST:-http://192.168.1.44:11434}"
TASKS_DIR="$(cd "$(dirname "$0")" && pwd)/tasks"
SCRIPTS_DIR="$(cd "$(dirname "$0")" && pwd)/scripts"

# ── Colors ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'

log()  { echo -e "${CYAN}▸${RESET} $*"; }
ok()   { echo -e "${GREEN}✓${RESET} $*"; }
warn() { echo -e "${YELLOW}⚠${RESET} $*"; }
err()  { echo -e "${RED}✗${RESET} $*" >&2; }

# ── Check dependencies ───────────────────────────────────────────────────────
check_deps() {
  local missing=()
  command -v claude &>/dev/null || missing+=("claude (Claude Code CLI)")
  command -v git    &>/dev/null || missing+=("git")
  command -v curl   &>/dev/null || missing+=("curl")
  command -v jq     &>/dev/null || missing+=("jq")

  if [[ ${#missing[@]} -gt 0 ]]; then
    err "Missing dependencies:"
    for dep in "${missing[@]}"; do echo "  - $dep"; done
    exit 1
  fi
}

# ── Ollama health check ──────────────────────────────────────────────────────
check_ollama() {
  log "Checking Ollama at $OLLAMA_HOST..."
  if ! curl -sf "$OLLAMA_HOST/api/tags" &>/dev/null; then
    err "Ollama not reachable at $OLLAMA_HOST"
    err "Make sure your Pi is on and Ollama is running."
    exit 1
  fi
  ok "Ollama reachable"
}

# ── Parse task frontmatter ───────────────────────────────────────────────────
get_field() {
  local file="$1" field="$2"
  grep "^${field}:" "$file" | head -1 | sed 's/^[^:]*: *//' | tr -d '"'
}

# ── List tasks ───────────────────────────────────────────────────────────────
list_tasks() {
  echo -e "\n${BOLD}Tasks:${RESET}\n"
  local found=0
  for f in "$TASKS_DIR"/task-[0-9]*.md; do
    [[ -f "$f" ]] || continue
    [[ "$f" == *"template"* ]] && continue
    local id title status
    id=$(get_field "$f" "id")
    title=$(get_field "$f" "title")
    status=$(get_field "$f" "status")
    local color="$RESET"
    case "$status" in
      pending)   color="$YELLOW" ;;
      done)      color="$GREEN"  ;;
      failed)    color="$RED"    ;;
      in-progress) color="$CYAN" ;;
    esac
    printf "  ${BOLD}%s${RESET}  ${color}%-12s${RESET}  %s\n" "$id" "$status" "$title"
    found=1
  done
  [[ $found -eq 0 ]] && warn "No tasks found in $TASKS_DIR"
  echo
}

# ── Pick task interactively ──────────────────────────────────────────────────
pick_task() {
  list_tasks
  local pending=()
  for f in "$TASKS_DIR"/task-[0-9]*.md; do
    [[ -f "$f" ]] || continue
    [[ "$f" == *"template"* ]] && continue
    local status
    status=$(get_field "$f" "status")
    [[ "$status" == "pending" ]] && pending+=("$f")
  done

  if [[ ${#pending[@]} -eq 0 ]]; then
    warn "No pending tasks. Create a task first with: ./scripts/new-task.sh"
    exit 0
  fi

  echo -e "${BOLD}Pending task IDs:${RESET}"
  for f in "${pending[@]}"; do
    local id title
    id=$(get_field "$f" "id")
    title=$(get_field "$f" "title")
    echo "  $id — $title"
  done
  echo
  read -rp "Enter task ID to run (or 'q' to quit): " choice
  [[ "$choice" == "q" ]] && exit 0
  echo "$choice"
}

# ── Run a task ───────────────────────────────────────────────────────────────
run_task() {
  local task_id="$1"
  local task_file
  task_file=$(find "$TASKS_DIR" -name "task-${task_id}-*.md" ! -name "*template*" | head -1)

  if [[ -z "$task_file" ]]; then
    err "No task file found for ID: $task_id"
    exit 1
  fi

  local title branch model
  title=$(get_field "$task_file" "title")
  branch=$(get_field "$task_file" "branch")
  model=$(get_field "$task_file" "model")

  echo -e "\n${BOLD}Running task ${task_id}:${RESET} $title"
  echo -e "  Branch : $branch"
  echo -e "  Model  : $model @ $OLLAMA_HOST\n"

  # Update status to in-progress
  sed -i '' "s/^status: .*/status: \"in-progress\"/" "$task_file" 2>/dev/null || \
  sed -i    "s/^status: .*/status: \"in-progress\"/" "$task_file"

  # Create and switch to task branch
  git checkout main 2>/dev/null || git checkout master
  git pull --ff-only
  git checkout -B "$branch"

  log "Launching Claude Code with local Ollama model: $model"

  # Build the prompt Claude Code will execute
  local prompt
  prompt="Read the task file at ${task_file}. Follow CLAUDE.md conventions exactly.
Use the Ollama model at ${OLLAMA_HOST} for code generation.
Complete all acceptance criteria, write the specified files, run any available tests.
When done, git add and commit with message matching the task ID format from CLAUDE.md.
Do not push — just commit locally."

  # Run Claude Code pointed at local Ollama
  ANTHROPIC_BASE_URL="$OLLAMA_HOST" \
  ANTHROPIC_AUTH_TOKEN="ollama" \
  ANTHROPIC_API_KEY="" \
    claude --model "$model" \
           --dangerously-skip-permissions \
           -p "$prompt"

  # Update status
  sed -i '' "s/^status: .*/status: \"done\"/" "$task_file" 2>/dev/null || \
  sed -i    "s/^status: .*/status: \"done\"/" "$task_file"

  ok "Task $task_id complete. Branch: $branch"
  echo
  log "Next: run './factory.sh review $task_id' or push the branch for a PR."
}

# ── Main ─────────────────────────────────────────────────────────────────────
main() {
  check_deps
  check_ollama

  case "${1:-}" in
    list)
      list_tasks
      ;;
    review)
      bash "$SCRIPTS_DIR/review.sh" "${2:-}"
      ;;
    "")
      task_id=$(pick_task)
      run_task "$task_id"
      ;;
    [0-9]*)
      run_task "$1"
      ;;
    *)
      echo "Usage: $0 [list | <task-id> | review <task-id>]"
      exit 1
      ;;
  esac
}

main "$@"
