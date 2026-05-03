---
name: task
description: Create a new task from a brief description and add it to the roadmap
argument-hint: "[description of what the task should do]"
---

Create a new task from $ARGUMENTS:

1. Read `docs/paas/ui/ROADMAP.md` to find the current highest task number and which milestone is active.
2. Determine the task type from the description: UI page/component → `docs/paas/ui/tasks/`, backend API → `docs/paas/backend/tasks/`.
3. Assign the next task number and a short slug (e.g. `M7-T32-vcn-peer-wiring.md`).
4. Create the task file using the template at `.claude/factory/tasks/task-000-template.md`.
   - Goal: one sentence from the description.
   - Context: infer from the description and current codebase state.
   - Acceptance criteria: brief testable bullets — do NOT generate implementation code or file content.
   - Files to create/modify: list paths only.
5. Add a one-line entry to the correct milestone table in `docs/paas/ui/ROADMAP.md` (or the backend roadmap if applicable).
6. Report: task file path, task number, and a one-line summary of what it covers.