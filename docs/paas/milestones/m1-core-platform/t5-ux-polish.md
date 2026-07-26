# T5 — UX Polish

## Goal

The platform feels production-grade: every loading, error, and empty state is handled and the user is never left guessing.

## Tasks

### 5.1 Loading states

- [x] Skeleton loaders for every list and detail view (ServiceInfoPage, SmerdPage — DeployPage/NewServicePage/VervClosedNetworkPage have no async load on mount, N/A)
- [x] Loading indicator in the page header during any in-flight API call (`TopBar.tsx` via `useIsFetching`)

### 5.2 Error boundaries

- [x] Page-level error boundary (`ErrorBoundary.tsx`)
- [x] API error states via `catchGrpc` toast
- [x] Retry button on failed data-fetching queries (`QueryErrorState`, wired into HomePage/ServiceInfoPage/SmerdPage)

### 5.3 Empty states

- [x] Empty-state component on HomePage for no services/smerds (with "Create a service" CTA)
- [x] Empty state for no deployments on ServiceInfoPage (also fixed a bug: this previously rendered fabricated mock deployment rows instead of an empty state)
- [x] Empty state for smerds list if empty (already done on HomePage; `SmerdsWidget.tsx` found to be an unused duplicate, flagged for a separate cleanup, not touched)

### 5.4 Toast / notification system

- [x] `bake` with success / warning / error variants
- [x] Auto-dismiss after 5 s; manual dismiss available
- [x] `catchGrpc` used in API `.catch()` blocks

### 5.5 Navigation & layout cleanup

- [x] Breadcrumb trail (`BreadcrumbsBar`, extracted to a shared component; ServiceInfoPage and SmerdPage both use it now — a true 3-level Home→Service→Deployment trail is N/A, there's no deployment-detail route yet)
- [x] Active route highlight in sidebar (`MainLayout.tsx`'s `ROUTE_TO_NAV` derives `activeNav` from `useLocation`, passed into `Sidebar`)
- [x] Responsive layout down to 1280 px wide (hamburger toggle in `TopBar`, `Sidebar` becomes a slide-in overlay with backdrop below 1280px; desktop behavior unchanged above it)

## Acceptance criteria

- [x] No page shows a blank white screen in known error cases
- [x] All list pages handle zero-item state with a message
- [x] Toast appears for every API failure
