# T2 — Deployments

## Goal

Let the operator see the deployment history for a service and trigger new deployments from the UI.

## Tasks

### 2.1 Deployments list (inside ServiceInfoPage)

- [x] Fetch deployments via `ListDeployments` filtered by service key
- [ ] Display: version/tag, deploy date, status, triggered-by — version/tag + triggered-by missing
- [x] Highlight the current (most-recent successful) deployment
- [ ] Paginate or virtualize if the list grows large

### 2.2 Deployment detail drawer / page

- [ ] Expand row to show image digest, env overrides, pipeline steps and outcomes — only spec ID, created-at, raw JSON
- [x] Collapsible raw JSON section for full deploy record

### 2.3 Deploy flow (DeployPage / DeployMenu)

- [x] Form exists via `DeploymentWidget` in `DeployMenu` (opens from ServiceInfoPage)
- [x] Submit calls `CreateNewDeployment` (service deployment API)
- [ ] Live status polling after submission — no step progress display
- [ ] Success: redirect to ServiceInfoPage with new deployment highlighted
- [ ] Failure: surface the failed step and error message via toast

### 2.4 "Deploy latest" shortcut

- [ ] One-click button on ServiceInfoPage — "Upgrade" tab in DeployMenu is "Not implemented yet"
- [ ] Confirmation dialog before submission

## Acceptance criteria

- [ ] Deploy form validates required fields before sending
- [ ] Pipeline progress is visible in real time (polling or streaming)
- [x] A failed deploy does not silently succeed — catchGrpc used in .catch()
