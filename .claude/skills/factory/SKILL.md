---
name: factory
description: Run a factory task using the Haiku→Ollama pipeline
argument-hint: "[task-file-path]"
---

Run the task using the factory pipeline:

1. If $ARGUMENTS is provided, treat it as the task file path. Otherwise, pick the next open task from
   `docs/paas/ui/tasks/` (lowest milestone, lowest task number with unchecked acceptance criteria items).

2. Spawn a **general-purpose** sub-agent with **model: haiku** and this prompt (replace `<resolved-task-file-path>`):

```
You are a factory worker agent. Execute the task spec file end-to-end without stopping to ask questions.

## Inputs
Task file: <resolved-task-file-path>
Ollama host: http://192.168.1.44:11434
Repo root: /Users/alexbukov/verv/Velez

## Workflow

1. Read the task file. Extract: Goal, Context, Acceptance Criteria, Files to create/modify.
2. Read existing files you will modify (always read before editing).
3. Call Ollama for each non-trivial code block:
   - UI/frontend (.tsx, .css) → model `stable-code`
   - Go backend (.go, .proto) → model `qwen2.5-coder:3b`
   ```bash
   curl -s http://192.168.1.44:11434/api/generate \
     -d '{"model":"stable-code","prompt":"<your prompt>","stream":false}' \
     | python3 -c "import sys,json; print(json.load(sys.stdin)['response'])"
   ```

4. Write/edit files using the generated output. Apply project rules below.
5. Verify:
   - UI task → `cd pkg/web/Velez-UI && yarn build:ui`
   - Backend task → `go test ./...` from repo root
   - Fix all TypeScript/Go errors before reporting done.

## Project rules

- Named function declarations only — `function Foo() {}`, never `const Foo = () => {}`
- Named handlers — event handlers are named function declarations
- CSS Modules — all styles in `.module.css`, no inline `style={{}}`
- Root class suffix `Container`; wrapper classes use `Wrapper`
- `@/` resolves to `src/` — use it for all imports
- `rem` for font sizes and spacing, never `px` for spacing
- No `!important`, no `z-index` hardcoding
- CSS nesting for child selectors inside a module
- No comments unless the WHY is non-obvious
- Go error handling: two separate lines — never `if err = foo(); err != nil`
- Go struct literals: assign to a named variable before passing to a function

```

3. The sub-agent handles everything end-to-end: reads the spec, calls Ollama for code generation,
   writes files, runs build/tests, commits, and reports back.

4. After the sub-agent returns, relay its report to the user: files changed, build result, commit hash,
   and any acceptance criteria that could not be met.

5. After the task is done mark it done in roadmap and milestones by checking the checkbox or adding strikethrough.