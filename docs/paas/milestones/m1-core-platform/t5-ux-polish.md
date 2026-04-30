# T5 — UX Polish

## Goal

The platform feels production-grade: every loading, error, and empty state is handled and the user is never left guessing.

## Tasks

### 5.1 Loading states
- Skeleton loaders for every list and detail view (not spinners — skeleton cards matching the layout)
- Loading indicator in the page header during any in-flight API call

### 5.2 Error boundaries
- Page-level error boundary catching unexpected render errors
- API error states: surface the gRPC error code and message via toast using `catchGrpc`
- Retry button on failed data-fetching queries

### 5.3 Empty states
- Custom empty-state component for: no services, no deployments, no smerds
- Include a primary CTA (e.g. "Create your first service") when applicable

### 5.4 Toast / notification system
- Audit all existing `bake` calls for consistency — success, warning, error variants
- Auto-dismiss after 5 s; manual dismiss available
- Ensure `catchGrpc` is called in every `.catch()` on API calls

### 5.5 Navigation & layout cleanup
- Breadcrumb trail: Home → Service → Deployment
- Active route highlight in any nav sidebar / header
- Responsive layout down to 1280 px wide (minimum supported width for an ops tool)

## Acceptance criteria

- No page ever shows a blank white screen in the known error cases
- All list pages handle zero-item state with a message
- Toast appears for every API failure, including network timeouts
