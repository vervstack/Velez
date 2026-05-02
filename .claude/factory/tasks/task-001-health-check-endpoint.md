---
id: "001"
title: "Health check endpoint"
status: "pending"
model: "qwen2.5-coder:3b"
created: "2026-05-02"
branch: "task/001-health-check-endpoint"
---

# Task 001 — Health check endpoint

## Goal

Create a `/health` HTTP endpoint in Go that returns server status as JSON.

## Context

This is the first endpoint in the project. The Go server entry point is `src/server/main.go`.
The handler should live in `src/server/handlers/health.go`.

## Acceptance Criteria

- [ ] GET `/health` returns HTTP 200 with `{"status":"ok","version":"0.1.0"}`
- [ ] Handler is registered in `main.go`
- [ ] Unit test covers happy path and verifies JSON shape
- [ ] Tests pass with `go test ./...`

## Files to Create / Modify

- `src/server/handlers/health.go` — new handler
- `src/server/handlers/health_test.go` — new test
- `src/server/main.go` — register the route (create if missing)

## Do NOT change

- Any existing files not listed above

## Notes

- Use `encoding/json` from stdlib, no external packages
- Version string can be hardcoded as `"0.1.0"` for now
