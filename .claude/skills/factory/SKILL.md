---
name: factory
description: Run a factory task using the Haiku→Ollama pipeline
argument-hint: "[task-file-or-description]"
---

Run the task using the factory pipeline:

1. If $ARGUMENTS is provided, treat it as the task file path or description. Otherwise, pick the next open task from
   `docs/paas/ui/tasks/` (lowest milestone, lowest task number with unchecked items).
2. Spawn a **Haiku sub-agent** for medium UI tasks (multi-file components, needs `yarn build:ui` check).
   Spawn a **Sonnet sub-agent** for complex cross-layer backend tasks (needs `go test ./...`).
3. The sub-agent must call Ollama at `http://192.168.1.44:11434` via `curl` for code generation:
   - UI/frontend code → model `stable-code`
   - Go backend code → model `qwen2.5-coder:3b`
4. Sub-agent writes generated output to the correct files, then runs the appropriate build/test command to verify no new
   errors were introduced.
5. Report back: what was done, what files changed, and any blockers (e.g. missing backend API).

Follow all rules in `CLAUDE.md` and `pkg/web/Velez-UI/CLAUDE.md` (named function declarations, CSS Modules, layer
imports, etc.).