# T2 — Deployments

## Goal

Let the operator see the deployment history for a service and trigger new deployments from the UI.

## Tasks

### 2.1 Deployments list (inside ServiceInfoPage)
- Fetch deployments via `ListDeployments` filtered by service key
- Display: version/tag, deploy date, status, triggered-by
- Highlight the current (most-recent successful) deployment
- Paginate or virtualize if the list grows large

### 2.2 Deployment detail drawer / page
- Expand a deployment row to show: image digest, env overrides, pipeline steps and their outcomes
- Collapsible raw JSON section for full deploy record

### 2.3 Deploy flow (DeployPage)
- Form: service selector, image tag input, optional env overrides
- Submit calls `CreateDeploy`
- Live status polling after submission — show pipeline step progress as steps complete
- Success: redirect to ServiceInfoPage with the new deployment highlighted
- Failure: surface the failed step and error message via toast

### 2.4 "Deploy latest" shortcut
- One-click button on ServiceInfoPage that re-deploys the current image tag
- Confirmation dialog before submission

## Acceptance criteria

- Deploy form validates required fields before sending
- Pipeline progress is visible in real time (polling or streaming)
- A failed deploy does not silently succeed — error is always surfaced
