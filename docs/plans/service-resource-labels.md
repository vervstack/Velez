# Service / resource classification + labels

## Goal

- Velez itself, and other Verv infra services, show up on the service page but are
  **filtered out by default** ("non-verv" / app-only view).
- Postgres (statefull mode) is modelled as a **resource bound to the `velez` service**,
  not as a service and not as a plugin-only concept.
- A lightweight label vocabulary drives UI badges + the filter.

## Decisions (locked)

1. **No labels table.** Labels are UI-only, **derived in the API response**, never persisted.
2. **Resource-vs-service detection = presence in `velez.service_resources`** (cluster) /
   `<service>_<x>` sibling-container naming (single-node). Not a label.
3. Resources are represented in **both** `velez.services` and `velez.service_resources`;
   the service list hides them via presence in `service_resources`.
4. Single-node keeps inferring resources from container names; provisioning just creates
   correctly-named containers. `local_storage` resource write stays a no-op.
5. `velez` in the service list:
   - **cluster:** seeded `velez.services` row (migration).
   - **single-node:** `local_storage` injects a synthetic entry, **deduplicated by name**
     against the container-derived list (the Velez container may already carry a
     user-set `VERV_SERVICE`).
   - Docker cannot change labels on a running container, so no label-on-self attempt.

## Label vocabulary (derived, not stored)

| Label | Rule |
|---|---|
| `resource-<type>` | name appears as `resource_name` in `service_resources` (cluster) / has `<svc>_<x>` sibling (single-node); `<type>` = `resource_type` |
| `service-core` | name in `domain.CoreServiceNames`, or (cluster) linked via `velez.plugins.service_id` |
| `service-app` | everything else |

`internal` (for the filter) = `service-core` OR `resource-*`.

`domain.CoreServiceNames`: `velez, matreshka, makosh, headscale, portainer, angie`.

## Contract

```proto
// ServiceBaseInfo
repeated string labels = 7;          // derived server-side

// ListServices.Request
bool include_internal = 3;           // default false -> hides core + resources
```

## Layers

### Backend

- **proto** — `ServiceBaseInfo.labels`, `ListServices.Request.include_internal`; `make codegen`.
- **domain** — `ServiceBaseInfo.Labels []string`, `ListServicesReq.IncludeInternal bool`,
  `CoreServiceNames` set, `ClassifyService(name string, resourceType string) []string`.
- **migration** — seed `velez` + `postgres` rows into `velez.services` (`ON CONFLICT DO NOTHING`).
- **storage/postgres** — `List`: derive `Labels`, apply `include_internal` filter in SQL
  (name NOT IN distinct `service_resources.resource_name`, name NOT IN core set, not plugin-linked).
  New sqlc query `ListResourceNames`.
- **storage/local_storage** — `listDistinctServices`: inject synthetic `velez` (dedupe by name),
  derive `Labels`, apply `include_internal` filter in-memory.
- **jobs/enable_statefull.go** — append job: upsert `service_resources('velez','postgres','postgres')`
  (first real caller of the existing `UpsertServiceResource` query).
- **transport/service_api_impl** — map `domain … Labels` -> proto `labels`; read `include_internal`.

### Frontend

- **processes/api** service list call — pass `include_internal`.
- **service list page** — filter toggle ("Show Verv internal"); per-row badge from `labels`.
- **widgets/service/ResourcesSection + ResourceCard** — drop the mock, wire to real
  `GetServiceResources`; render on the `velez` service page with Postgres as a bound resource.

## Slices / order

1. proto + domain + migration (sequential, blocks everything).
2. storage (pg + local_storage) -> service/transport wiring -> `enable_statefull` job. *(Sonnet)*
3. frontend: api param, filter control, badges, ResourcesSection real data. *(Haiku, parallel to 2 once 1 lands)*

## Not in scope

- Removing `statefull_pg` from `velez.plugins` / the plugin lifecycle UI (additive change only).
- Object storage / new resource kinds. "Docker-based stateless storage" here just means the
  existing `local_storage` backend, which must support the same scheme as `storage/postgres`.
