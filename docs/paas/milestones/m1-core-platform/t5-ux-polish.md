# T5 — UX Polish

## Goal

The platform feels production-grade: every loading, error, and empty state is handled and the user is never left guessing.

## Tasks

### 5.1 Loading states

- [ ] Skeleton loaders for every list and detail view — currently plain "Loading..." text only
- [ ] Loading indicator in the page header during any in-flight API call

### 5.2 Error boundaries

- [x] Page-level error boundary (`ErrorBoundary.tsx`)
- [x] API error states via `catchGrpc` toast
- [ ] Retry button on failed data-fetching queries

### 5.3 Empty states

- [x] Empty-state component on HomePage for no services/smerds (with "Create a service" CTA)
- [ ] Empty state for no deployments on ServiceInfoPage
- [ ] Empty state for smerds list if empty

### 5.4 Toast / notification system

- [x] `bake` with success / warning / error variants
- [x] Auto-dismiss after 5 s; manual dismiss available
- [x] `catchGrpc` used in API `.catch()` blocks

### 5.5 Navigation & layout cleanup

- [ ] Breadcrumb trail: Home → Service → Deployment (currently plain "← Back" links)
- [ ] Active route highlight in sidebar
- [ ] Responsive layout down to 1280 px wide

## Acceptance criteria

- [x] No page shows a blank white screen in known error cases
- [ ] All list pages handle zero-item state with a message
- [x] Toast appears for every API failure
