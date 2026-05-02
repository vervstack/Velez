#!/usr/bin/env bash
# scripts/status.sh — show factory pipeline at a glance

TASKS_DIR="$(cd "$(dirname "$0")/.." && pwd)/tasks"
REVIEWS_DIR="$(cd "$(dirname "$0")/.." && pwd)/reviews"

BOLD='\033[1m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
RED='\033[0;31m'; CYAN='\033[0;36m'; DIM='\033[2m'; RESET='\033[0m'

get_field() { grep "^${2}:" "$1" | head -1 | sed 's/^[^:]*: *//' | tr -d '"'; }

echo -e "\n${BOLD}Factory Pipeline Status${RESET}"
echo -e "${DIM}$(date)${RESET}\n"

printf "${BOLD}%-6s %-14s %-12s %-10s %s${RESET}\n" "ID" "STATUS" "REVIEW" "MODEL" "TITLE"
echo "──────────────────────────────────────────────────────────────────"

for f in "$TASKS_DIR"/task-[0-9]*.md; do
  [[ -f "$f" ]] || continue
  [[ "$f" == *"template"* ]] && continue

  id=$(get_field "$f" "id")
  title=$(get_field "$f" "title")
  status=$(get_field "$f" "status")
  model=$(get_field "$f" "model")

  # Check for review file
  review_file="$REVIEWS_DIR/review-${id}.md"
  review_verdict="—"
  if [[ -f "$review_file" ]]; then
    if grep -q "✅ PASS" "$review_file"; then
      review_verdict="${GREEN}✅ pass${RESET}"
    elif grep -q "⚠️ PARTIAL" "$review_file"; then
      review_verdict="${YELLOW}⚠️  partial${RESET}"
    elif grep -q "❌ FAIL" "$review_file"; then
      review_verdict="${RED}❌ fail${RESET}"
    else
      review_verdict="${DIM}reviewed${RESET}"
    fi
  fi

  status_color="$RESET"
  case "$status" in
    pending)     status_color="$YELLOW" ;;
    in-progress) status_color="$CYAN"   ;;
    done)        status_color="$GREEN"  ;;
    failed)      status_color="$RED"    ;;
  esac

  printf "%-6s ${status_color}%-14s${RESET} %-22b %-10s %s\n" \
    "$id" "$status" "$review_verdict" "${model%%:*}" "$title"
done

echo
