# Known Pitfalls

## Never normalize a value through a pointer field a caller owns

A request struct with a pointer field (`*velez_api.ListSmerds_Request.Name`) is often built as
`&callerOwnedField`, not a copy — e.g. `verv_services.List`/`.Get` build
`&velez_api.ListSmerds_Request{Name: &svc.Name}` pointing straight at their own
`domain.Service.Name`/`domain.ServiceBaseInfo.Name` field. `container_manager.ListSmerds` used to
lowercase its filter in place (`*req.Name = strings.ToLower(req.GetName())`) to normalize the
Docker name filter, which silently rewrote the caller's `svc.Name` too — every service name with an
uppercase letter (any satellite instance a caller names itself: runner, registry, pgaas) came back
from `ListServices` already lowercased, and the exact name the UI then fed back into `GetService`
no longer matched the real, case-sensitive `VervServiceLabel` on the container, 404ing a service
the list had just handed out (`not found:error getting service by name from storage:error
getting service info`). Fixed in `smerd_list.go` by building a fresh
`*velez_api.ListSmerds_Request` instead of mutating through the shared pointer. No linter catches
this — a pointer field being caller-owned vs. callee-owned isn't visible from the callee's
signature. Before mutating through any pointer field on an incoming request/response struct, copy
the value out (or build a new struct) instead. Regression coverage:
`tests/e2e/suite_service_list_case_test.go`.

## Docker's `name` filter is an unanchored substring match

Never use `ListSmerds_Request.Name` to select a service's containers (removing `dind` also matched
`dind-gitlab`); filter by the exact `Label: {VERV_SERVICE: name}` instead. `VervService.Remove`
also refuses (`ErrServiceHasDependents`, listing the dependents) while `VERV_DEPENDS_ON` callers or
runners on the service's DinD exist.
