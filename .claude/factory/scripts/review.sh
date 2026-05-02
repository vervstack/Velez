#!/usr/bin/env bash
# scripts/review.sh — send completed task branch to Haiku for review
# Usage: ./scripts/review.sh <task-id>
# Requires: ANTHROPIC_API_KEY set in environment (standard Anthropic key)

set -euo pipefail

TASKS_DIR="$(cd "$(dirname "$0")/.." && pwd)/tasks"
REVIEWS_DIR="$(cd "$(dirname "$0")/.." && pwd)/reviews"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'

log()  { echo -e "${CYAN}▸${RESET} $*"; }
ok()   { echo -e "${GREEN}✓${RESET} $*"; }
err()  { echo -e "${RED}✗${RESET} $*" >&2; }

get_field() {
  local file="$1" field="$2"
  grep "^${field}:" "$file" | head -1 | sed 's/^[^:]*: *//' | tr -d '"'
}

review_task() {
  local task_id="$1"
  local task_file
  task_file=$(find "$TASKS_DIR" -name "task-${task_id}-*.md" ! -name "*template*" | head -1)

  if [[ -z "$task_file" ]]; then
    err "No task file for ID: $task_id"
    exit 1
  fi

  if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
    err "ANTHROPIC_API_KEY not set. Export it before running review."
    exit 1
  fi

  local branch
  branch=$(get_field "$task_file" "branch")
  local task_content
  task_content=$(cat "$task_file")

  log "Getting diff for branch: $branch"
  local diff
  diff=$(git diff main..."$branch" -- . ':(exclude)*.md' 2>/dev/null || \
         git diff master..."$branch" -- . ':(exclude)*.md')

  if [[ -z "$diff" ]]; then
    err "No diff found between main and $branch"
    exit 1
  fi

  # Truncate diff if huge (Haiku 200k context, but keep cost low)
  local diff_lines
  diff_lines=$(echo "$diff" | wc -l)
  if [[ $diff_lines -gt 800 ]]; then
    log "Diff is large ($diff_lines lines), truncating to first 800 lines"
    diff=$(echo "$diff" | head -800)
    diff="${diff}

[... diff truncated for review cost ...]"
  fi

  log "Sending to Claude Haiku for review..."

  local prompt
  prompt="You are a code reviewer. Be concise and direct.

TASK SPEC:
---
${task_content}
---

GIT DIFF (what the agent produced):
---
${diff}
---

Review the diff against the task spec. Respond in this exact format:

VERDICT: ✅ PASS | ⚠️ PARTIAL | ❌ FAIL

SUMMARY: One sentence.

CRITERIA:
- [✅/❌] Each acceptance criterion from the task, checked one by one

ISSUES:
- List any bugs, missing files, style violations, or spec mismatches (or 'None')

SUGGESTION:
One concrete next step if verdict is not ✅, or 'Ready to merge' if it is."

  # Call Haiku API
  local response
  response=$(curl -sf "https://api.anthropic.com/v1/messages" \
    -H "x-api-key: $ANTHROPIC_API_KEY" \
    -H "anthropic-version: 2023-06-01" \
    -H "content-type: application/json" \
    -d "$(jq -n \
      --arg model "claude-haiku-4-5-20251001" \
      --arg prompt "$prompt" \
      '{model: $model, max_tokens: 1024, messages: [{role: "user", content: $prompt}]}')")

  local review_text
  review_text=$(echo "$response" | jq -r '.content[0].text')

  if [[ -z "$review_text" || "$review_text" == "null" ]]; then
    err "Empty response from Haiku. Raw response:"
    echo "$response"
    exit 1
  fi

  # Write review file
  mkdir -p "$REVIEWS_DIR"
  local review_file="$REVIEWS_DIR/review-${task_id}.md"
  cat > "$review_file" <<EOF
# Review — Task ${task_id}

**Branch**: \`${branch}\`
**Reviewed**: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
**Model**: claude-haiku-4-5-20251001

---

${review_text}
EOF

  echo
  cat "$review_file"
  echo
  ok "Review saved to $review_file"

  # Show next step based on verdict
  if echo "$review_text" | grep -q "✅ PASS"; then
    echo -e "\n${GREEN}${BOLD}Ready to push and open PR:${RESET}"
    echo "  git push -u origin $branch"
    echo "  gh pr create --title '[${task_id}] $(get_field "$task_file" "title")' --body-file $review_file"
  else
    echo -e "\n${YELLOW}${BOLD}Fix needed. Re-run the task or edit manually, then:${RESET}"
    echo "  ./factory.sh $task_id"
  fi
}

if [[ -z "${1:-}" ]]; then
  err "Usage: $0 <task-id>"
  exit 1
fi

review_task "$1"
