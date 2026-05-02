# Velez Coding Factory — Integration Plan

## What is This

A local agent pipeline that turns task files into committed code branches, reviewed and ready to merge.
You define the task. The factory executes it. You merge or reject.

The factory already exists in `.claude/factory/`. This document is about what works, what doesn't, and how to
actually wire it into the Velez project — covering both the frontend UI redesign backlog and the backend PaaS work.

---

## Current State Assessment

### What's good

- Task format (`task-NNN-*.md` with frontmatter) is clean and self-contained.
- Haiku review step (`scripts/review.sh`) works as-is — it just needs `ANTHROPIC_API_KEY`.
- `scripts/status.sh` and `scripts/new-task.sh` are ready to use.
- The task backlog already exists: 29 UI tasks in `docs/paas/ui/tasks/`, 5 milestone areas in `docs/paas/`.

### What's broken / misleading

The core execution step in `factory.sh` won't work:

```bash
ANTHROPIC_BASE_URL="$OLLAMA_HOST" \
ANTHROPIC_AUTH_TOKEN="ollama" \
ANTHROPIC_API_KEY="" \
  claude --model "$model" ...
```

Claude Code CLI speaks **Anthropic Messages API format**. Ollama's default API (`/api/generate`, `/api/chat`)
speaks a different format. Setting `ANTHROPIC_BASE_URL` to Ollama's host will cause 404s or malformed responses.

Ollama does expose an OpenAI-compatible endpoint at `/v1/chat/completions`, but Claude Code CLI doesn't speak
OpenAI format either — it speaks Anthropic format only.

**Nothing is broken that can't be fixed.** But the execution layer needs to be replaced.

---

## Architecture: Two Viable Approaches

### Option A — Claude Code as Orchestrator + Anthropic Sub-agents (Recommended for complex tasks)

```
You  →  Claude Code (Sonnet, this conversation)
           ├─ reads task file
           ├─ spawns sub-agent via Agent tool  →  Haiku or Sonnet (Anthropic infra)
           │       └─ writes code, runs tests, commits
           └─ Haiku reviews diff
```

**How sub-agents work on my side:**
I have an `Agent` tool. When you tell me to run a task, I can spawn a sub-agent that has the same tools
(Read, Write, Edit, Bash, etc.) and works inside the same repo. The sub-agent is isolated — it gets a
written prompt, does its work, and returns a result. I see the result; it does not persist state.

This is **fully on my side** — you don't configure anything for it. Sub-agents run on Anthropic's infrastructure,
cost money per task (Haiku ≈ $0.01–0.03, Sonnet ≈ $0.10–0.30 depending on task size).

**What you provide:**
- Nothing extra. Your Anthropic subscription or API key covers it.

**Best for:** Multi-file tasks, tasks that need to run tests, tasks with complex context requirements.
All 29 UI tasks in `docs/paas/ui/tasks/` fall here — they need to read existing components, run `yarn build`,
check TypeScript errors.

---

### Option B — Shell Factory with Direct Ollama API Calls (Recommended for bulk/template generation)

Replace the broken `claude` CLI call in `factory.sh` with a direct Ollama API call:

```bash
# Instead of the broken claude CLI trick:
curl -sf "$OLLAMA_HOST/api/generate" \
  -d "$(jq -n --arg model "$model" --arg prompt "$prompt" \
    '{model: $model, prompt: $prompt, stream: false}')" \
  | jq -r '.response' > /tmp/agent-output.txt
```

The agent's output is raw text (code blocks, file contents). You need a thin parser that extracts file
contents from the response and writes them to disk.

```
factory.sh
  └─ builds context prompt (task file + relevant source files)
  └─ calls Ollama /api/generate
  └─ parses response → writes files
  └─ runs tests (go test ./... or yarn build)
  └─ commits if tests pass
  └─ review.sh sends diff to Haiku
```

**What you provide:**
- `OLLAMA_HOST=http://192.168.1.44:11434` (already in .env.example)
- Frontend model name (run `curl http://192.168.1.44:11434/api/tags | jq '.models[].name'` to list them)
- `ANTHROPIC_API_KEY` for the Haiku review step

**Best for:** Self-contained, single-file tasks. Design token CSS, simple components, boilerplate. Tasks where
the output is predictable and the model just needs to fill in a template.

**Limitation:** Ollama models (3B) can't reliably multi-step (read file → understand → write related file).
They do one shot. Complex tasks will produce incomplete or incorrect output. Plan for 30–50% review failure
rate and budget reruns.

---

### Recommended Setup: Hybrid

Use both approaches based on task complexity:

| Task type | Execution | Model | Cost |
|-----------|-----------|-------|------|
| Simple component, single file | Ollama direct (Option B) | frontend model | free |
| Multi-file, needs context | Claude sub-agent (Option A) | Haiku | ~$0.02 |
| Complex logic, test-driven | Claude sub-agent (Option A) | Sonnet | ~$0.20 |
| Code review | Haiku API | Haiku | ~$0.02 |

Add a `complexity` field to the task frontmatter: `simple` → Ollama, `medium`/`complex` → Anthropic sub-agent.

---

## What You Need

### Required

| Item | Where | Status |
|------|-------|--------|
| `ANTHROPIC_API_KEY` | console.anthropic.com | Needed for Haiku reviews |
| `OLLAMA_HOST` | `http://192.168.1.44:11434` | Already in .env.example |
| Frontend model name | run `curl $OLLAMA_HOST/api/tags` | Confirm exact name |

### Optional

| Item | Purpose |
|------|---------|
| `litellm` proxy on the Pi | Makes Ollama look like Anthropic API — enables the original `claude` CLI approach if you want it |
| `gh` CLI configured | Auto-open PRs after review passes |

### For Claude Code sub-agent mode (Option A)

Nothing extra needed. Sub-agent spawning is built into Claude Code. You just tell me "run task T02" and I
do it. I read the task file, spawn the sub-agent with a contextualized prompt, monitor it, and report back.

---

## Integration with Velez Tasks

### UI Redesign (docs/paas/ui/) — 29 tasks

These are the best candidates to run through the factory first. They are:

- **Self-contained** — each task specifies exact files to create/modify
- **Ordered** — M1 → M2 → ... (can't skip milestones)
- **Testable** — `yarn build` catches TypeScript errors
- **Well-specified** — acceptance criteria are clear

Recommended execution order:
1. **T01 (design tokens)** — run manually first to establish the baseline CSS variables
2. **T02–T09 (base components)** — all can run in parallel (no interdependencies within M2)
3. **T10–T15 (complex components)** — run after M2 is merged
4. Continue milestone by milestone

For parallel execution I can spawn multiple sub-agents in one turn — all 8 M2 tasks simultaneously.

### Backend PaaS (docs/paas/milestones/) — M1–M5

These need more care. Backend tasks touch Go code, migrations, and the gRPC API. Use Sonnet sub-agents, not
Ollama, for these. Factory tasks for backend should:
- Reference specific proto files and Go layer boundaries
- Include `go test ./...` as an acceptance criterion
- Be scoped to one layer at a time (don't cross transport → service → storage in one task)

---

## Step-by-Step: Getting Started Today

### Step 1 — Verify Ollama

```bash
# List available models on your Pi
curl http://192.168.1.44:11434/api/tags | jq '.models[].name'
```

Confirm the frontend model name and update `.claude/factory/.env.example` with it.

### Step 2 — Set up credentials

```bash
cp .claude/factory/.env.example .claude/factory/.env
# Edit .env:
#   ANTHROPIC_API_KEY=sk-ant-...   ← your real key
#   OLLAMA_HOST=http://192.168.1.44:11434
#   FRONTEND_MODEL=<name from step 1>
```

`.claude/factory/.env` is already in `.gitignore` — safe to put the key there.

### Step 3 — Fix factory.sh execution layer

The current `run_task()` function needs to be split based on complexity:

```bash
# In run_task(), replace the broken claude CLI call with:
if [[ "$complexity" == "simple" ]]; then
  # Direct Ollama call → parse output → write files
  bash "$SCRIPTS_DIR/run-ollama.sh" "$task_file" "$model"
else
  # Hand off to Claude Code sub-agent orchestration (tell me: "run task NNN")
  echo "This task requires Claude Code orchestration."
  echo "Tell Claude Code: 'Run factory task $task_id from $task_file'"
fi
```

See `docs/factory/scripts/run-ollama.sh` for the implementation (to be written in Step 4).

### Step 4 — Write the Ollama execution script

Create `.claude/factory/scripts/run-ollama.sh` — a script that:
1. Reads the task file
2. Collects referenced source files as context
3. Builds a structured prompt
4. Calls `/api/generate`
5. Parses the response (extracts fenced code blocks)
6. Writes files to disk
7. Runs `yarn build` or `go test ./...`
8. Returns exit code based on test result

This is a ~100-line bash script. Tell me to write it when you're ready.

### Step 5 — Run your first task

Start with T01 (design tokens) — it's a single CSS file, zero dependencies, pure generation:

```bash
# Option A: Tell me directly
# "Run the factory task at docs/paas/ui/tasks/M1-T01-design-tokens.md"
# I'll read it, spawn a sub-agent, and commit the result.

# Option B: Shell factory (after Step 4 is done)
cd .claude/factory
source .env
./factory.sh 001   # or whatever ID you assign T01
```

### Step 6 — Review and merge

```bash
cd .claude/factory
./scripts/review.sh <task-id>
# If ✅ PASS:
git push -u origin <branch>
gh pr create --title "[T01] Design tokens"
```

---

## Task File Adaptation for Velez

The current task template (`task-NNN-*.md`) works but needs two additions for Velez:

```markdown
---
id: "T01"
title: "Design tokens"
status: "pending"
complexity: "simple"          # ← ADD: simple | medium | complex
model: "qwen2.5-coder:3b"    # for simple; ignored for medium/complex (uses Haiku/Sonnet)
created: "2026-05-02"
branch: "task/T01-design-tokens"
project: "ui"                  # ← ADD: ui | backend | infra
---
```

`complexity` drives which execution path to use. `project` maps to which test command to run.

---

## What I Need From You to Start

1. **Confirm the frontend model name** — run `curl http://192.168.1.44:11434/api/tags | jq '.models[].name'`
2. **Give me an `ANTHROPIC_API_KEY`** — needed for Haiku reviews. If you already have one in `.env`, we're set.
3. **Tell me which task to run first** — I'd recommend T01 (design tokens) as a dry run.

For sub-agent execution: nothing else needed. Tell me "run task T01" and I'll handle it from here using the
Agent tool. The sub-agent runs inside this repo with full tool access.

---

## Summary

| Question | Answer |
|----------|--------|
| Can you spawn sub-agents? | Yes — I do it via Agent tool, runs on Anthropic infra |
| Does the current factory.sh work? | Partially — review.sh works, run_task() is broken (Ollama API mismatch) |
| Do you need an API key? | Yes — for Haiku reviews. Sub-agents via me use your existing subscription |
| Do you need the Ollama endpoint? | Only for Option B (direct generation). Already known: 192.168.1.44:11434 |
| Best first task? | T01 design tokens — single file, zero deps, good smoke test |
| Can M2 tasks run in parallel? | Yes — I can spawn all 8 simultaneously via Agent tool |
