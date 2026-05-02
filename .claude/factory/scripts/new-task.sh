#!/usr/bin/env bash
# scripts/new-task.sh — create a new task file from template
# Usage: ./scripts/new-task.sh "Auth middleware handler" [model]

set -euo pipefail

TASKS_DIR="$(cd "$(dirname "$0")/.." && pwd)/tasks"

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; RESET='\033[0m'
ok()  { echo -e "${GREEN}✓${RESET} $*"; }
err() { echo -e "${RED}✗${RESET} $*" >&2; }
log() { echo -e "${CYAN}▸${RESET} $*"; }

if [[ -z "${1:-}" ]]; then
  err "Usage: $0 \"Task title\" [model]"
  err "  model defaults to qwen2.5-coder:3b"
  exit 1
fi

TITLE="$1"
MODEL="${2:-qwen2.5-coder:3b}"
DATE=$(date +%Y-%m-%d)

# Get next ID
last_id=$(ls "$TASKS_DIR"/task-[0-9]*.md 2>/dev/null | grep -v template | \
          sed 's/.*task-//' | sed 's/-.*//' | sort -n | tail -1)
next_id=$(printf "%03d" $(( ${last_id:-0} + 1 )))

# Slugify title
slug=$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$//')
filename="task-${next_id}-${slug}.md"
filepath="$TASKS_DIR/$filename"
branch="task/${next_id}-${slug}"

cat > "$filepath" <<EOF
---
id: "${next_id}"
title: "${TITLE}"
status: "pending"
model: "${MODEL}"
created: "${DATE}"
branch: "${branch}"
---

# Task ${next_id} — ${TITLE}

## Goal
<!-- One clear sentence describing what this task produces. -->

## Context
<!-- Background the agent needs. File paths, related components, API contracts. -->

## Acceptance Criteria
- [ ] 
- [ ] 
- [ ] Tests pass

## Files to Create / Modify
- \`src/\` — 

## Do NOT change
<!-- List files the agent must not touch -->

## Notes
<!-- Constraints, gotchas, references -->
EOF

ok "Created: $filepath"
log "Edit the task, then run: ./factory.sh $next_id"
echo
echo "  \$EDITOR $filepath"
