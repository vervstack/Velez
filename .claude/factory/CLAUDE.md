# Project Factory — Agent Context

## Architecture

- **Backend**: Go (standard library preferred, minimal deps)
- **Frontend**: Vanilla HTML/CSS/JS or lightweight framework (no heavy bundlers)
- **Pi Ollama**: http://192.168.1.44:11434 — used for code generation
- **Review**: Claude Haiku via Anthropic API

## Code Style

### Go

- Package names: short, lowercase, no underscores
- Error handling: always explicit, no silent ignores
- File naming: `snake_case.go`
- Tests alongside source: `handler_test.go` next to `handler.go`
- Prefer `net/http` stdlib over frameworks unless complexity demands otherwise

### Frontend

- No build step unless necessary — prefer single-file components
- CSS: BEM naming or utility classes, no CSS-in-JS
- JS: ES modules, no TypeScript unless project already uses it
- Accessibility: semantic HTML, aria labels on interactive elements

## Task System

- Tasks live in `tasks/` as markdown files
- Naming: `task-NNN-short-description.md` (e.g. `task-001-auth-handler.md`)
- Status tracked in task file frontmatter
- One branch per task: `task/NNN-short-description`
- After completion: commit, push branch, trigger Haiku review

## Git Conventions

- Branch from `main` for each task
- Commit message format: `[NNN] short description of what was done`
- Never commit directly to `main`
- PR description should include task file content + review result

## Agent Rules

1. Read the task file fully before writing any code
2. Ask no clarifying questions — infer from task + CLAUDE.md
3. Write tests for all Go handlers
4. After writing code, run available tests before committing
5. If a task is ambiguous, implement the simplest correct interpretation and note it in the review file
