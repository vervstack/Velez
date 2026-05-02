# 🏭 Coding Factory

Local agent pipeline: Claude Code picks tasks → Pi Ollama executes → Haiku reviews → you merge.

## Setup

```bash
# 1. Copy env file and fill in your API key
cp .env.example .env
$EDITOR .env

# 2. Make scripts executable
chmod +x factory.sh scripts/*.sh

# 3. Verify Ollama is reachable
curl http://192.168.1.44:11434/api/tags
```

## Daily Workflow

```bash
source .env

# See what's pending
./factory.sh list

# Create a new task (then edit it)
./scripts/new-task.sh "Login form UI"

# Run a task (interactive picker)
./factory.sh

# Run a specific task
./factory.sh 001

# Review completed task (sends diff to Haiku)
./factory.sh review 001

# Push branch + open PR (after ✅ review)
git push -u origin task/001-health-check-endpoint
gh pr create --title "[001] Health check endpoint"

# See full pipeline status
./scripts/status.sh
```

## Pipeline Architecture

```
You + Claude Sonnet (here or Claude Code)
  └─ writes tasks/task-NNN-*.md

factory.sh
  └─ creates git branch
  └─ runs Claude Code CLI
       └─ model: qwen2.5-coder:3b @ Pi (192.168.1.44:11434)
       └─ reads task file + CLAUDE.md
       └─ writes code, runs tests, commits

scripts/review.sh
  └─ diffs branch vs main
  └─ sends to Claude Haiku API
  └─ writes reviews/review-NNN.md
  └─ verdict: ✅ PASS / ⚠️ PARTIAL / ❌ FAIL

You
  └─ reads review, merges PR (or re-runs task)
```

## Task File Format

See `tasks/task-000-template.md`. Key fields:

- `model` — which Ollama model to use (`qwen2.5-coder:3b` or `stable-code:3b` for frontend)
- `status` — `pending` → `in-progress` → `done`
- `branch` — auto-used for git branch name

## Cost

- Code generation: **free** (local Ollama on Pi)
- Review per task: ~$0.02–0.05 (Claude Haiku)
- Planning: your Claude subscription (used here in chat)
