# T31 — Wire mock pages to real API (frontend)

**Status:** pending

## Goal

Replace all hardcoded `MOCK_*` arrays in the rebuilt pages with live React Query calls.
This task covers only what the existing API already supports. Fields not yet available from
the backend use safe defaults documented in each mapping.

## Blocked dependency

Before this task can be completed fully, **T32** (backend API enrichment) must be done.
The partial wiring below is still worth landing so real service names appear and the
skeleton is in place for T32 to flesh out.

---

## Pages and what to wire

### 1. DeploymentsPage (`src/pages/deployments/DeploymentsPage.tsx`)

Replace `MOCK_SERVICES: ServiceCardData[]` with a React Query call:

```ts
import {FetchSmerds} from "@/processes/api/velez";
import {useSettings} from "@/app/settings/state.ts";

const {initReq} = useSettings();
const smerdsQuery = useQuery({
    queryKey: ["smerds"],
    queryFn: () => FetchSmerds(initReq).catch((e) => { toaster.catchGrpc(e); return {smerds: []}; }),
    refetchInterval: 5000,
});
const services: ServiceCardData[] = (smerdsQuery.data?.smerds ?? []).map(mapSmerdToServiceCard);
```

Mapping function (add to `src/processes/mappings/smerds.ts`, create the file):

```ts
import {Smerd, SmerdStatus} from "@/app/api/velez";
import {ServiceCardData} from "@/components/service/ServiceCard";

export function mapSmerdToServiceCard(s: Smerd): ServiceCardData {
    return {
        name:          s.name ?? '',
        image:         s.imageName ?? '',
        status:        mapSmerdStatus(s.status),
        cpu:           0,      // not yet available from API — added in T32
        mem:           0,      // not yet available from API — added in T32
        uptime:        '-',    // not yet available from API — added in T32
        restarts:      0,      // not yet available from API — added in T32
        env:           s.labels?.['env'] ?? 'prod',
        incident:      false,
        releaseFrozen: false,
        node:          LOCAL_NODE,
    };
}

function mapSmerdStatus(s?: SmerdStatus): 'running' | 'degraded' | 'stopped' {
    switch (s) {
        case SmerdStatus.running:    return 'running';
        case SmerdStatus.restarting: return 'degraded';
        case SmerdStatus.dead:
        case SmerdStatus.exited:
        case SmerdStatus.paused:
        case SmerdStatus.removing:   return 'stopped';
        default:                     return 'stopped';
    }
}

const LOCAL_NODE = {id: 'local', host: 'localhost', status: 'online' as const};
```

Pass `services` where `MOCK_SERVICES` was used. Show a spinner or empty state while loading
(reuse the existing `EmptyState` component if present).

---

### 2. AppsPage (`src/pages/apps/AppsPage.tsx`)

Replace `MOCK_APPS: AppData[]` with the same `FetchSmerds` query + mapping function:

```ts
import {mapSmerdToAppData} from "@/processes/mappings/smerds";

// in smerds.ts — add alongside mapSmerdToServiceCard:
export function mapSmerdToAppData(s: Smerd): AppData {
    return {
        name:          s.name ?? '',
        image:         s.imageName ?? '',
        status:        mapSmerdStatus(s.status),
        env:           s.labels?.['env'] ?? 'prod',
        incident:      false,
        releaseFrozen: false,
        node:          LOCAL_NODE,
        deployments:   0,       // not yet available — added in T32
        lastDeployed:  '-',     // not yet available — added in T32
        configSource:  'none',
        version:       s.imageName?.split(':')[1] ?? 'latest',
    };
}
```

Replace both `MOCK_APPS` usages (data source + filter) with the query result.

---

### 3. SearchPage (`src/pages/search/SearchPage.tsx`)

Replace `MOCK_SERVICES` with the same `FetchSmerds` query. Replace `MOCK_NODES` with an
empty array `[]` — node search is blocked on a `ListNodes` API (tracked in T32).

The search filter logic that operates on `MOCK_SERVICES` should be unchanged; just swap the
data source.

---

### 4. ControlPlanePage (`src/pages/controlplane/ControlPlanePage.tsx`)

**Plugins section:** replace `MOCK_PLUGINS` with `ListVervServices()`.

`ListVervServices` returns `Service[]` where `Service` has `type: VervServiceType` and
`state: VervServiceState`. Map to the plugin display shape the page currently uses.

```ts
import {ListVervServices} from "@/processes/api/control_plane";
import {useSettings} from "@/app/settings/state.ts";
import {VervServiceState} from "@/app/api/velez";

const {initReq} = useSettings();
const pluginsQuery = useQuery({
    queryKey: ["vervServices"],
    queryFn: () => ListVervServices(initReq).catch((e) => { toaster.catchGrpc(e); return []; }),
    refetchInterval: 10000,
});
```

Map `VervServiceState` to the plugin status values the PluginMatrix widget expects. If those
values are strings like `'enabled' | 'disabled' | 'error'`, define the mapping locally in the page.

**Nodes section:** keep `MOCK_NODES` — a `ListNodes` endpoint does not yet exist.
Add a visible `// TODO(T32): replace with ListNodes once backend API is available` comment.

---

## Files to create

- `src/processes/mappings/smerds.ts` — `mapSmerdToServiceCard`, `mapSmerdToAppData`, `mapSmerdStatus`, `LOCAL_NODE`

## Files to modify

- `src/pages/deployments/DeploymentsPage.tsx` — replace MOCK_SERVICES with query
- `src/pages/apps/AppsPage.tsx` — replace MOCK_APPS with query
- `src/pages/search/SearchPage.tsx` — replace MOCK_SERVICES with query; MOCK_NODES → `[]`
- `src/pages/controlplane/ControlPlanePage.tsx` — replace MOCK_PLUGINS with query; mark MOCK_NODES TODO

## Do NOT change

- `src/components/service/ServiceCard.tsx` — card component must not change
- `src/components/apps/AppCard.tsx` — card component must not change
- `src/widgets/controlplane/PluginMatrix.tsx` — widget must not change
- `src/pages/vcn/VervClosedNetworkPage.tsx` — VCN peers need backend work first (T32)
- `src/app/router/MainLayout.tsx` — node selector needs backend work first (T32)

## Acceptance Criteria

- [ ] `DeploymentsPage` renders real container names and images from the backend
- [ ] `AppsPage` renders real container names and images from the backend
- [ ] `SearchPage` filters real container results; nodes section renders empty with no crash
- [ ] `ControlPlanePage` plugins section shows real installed services; nodes section still uses mock
- [ ] All four pages show a loading state while queries are in flight
- [ ] gRPC errors surface as toasts (via `toaster.catchGrpc`)
- [ ] `yarn build:ui` passes with no TypeScript errors
- [ ] No `MOCK_SERVICES` or `MOCK_APPS` or `MOCK_PLUGINS` constants remain in the wired pages

## Notes

- `FetchSmerds` already exists in `src/processes/api/velez.ts` — use it directly
- `useSettings` is the hook that provides `initReq` — look at `HomePage.tsx` for the exact pattern
- Keep `LOCAL_NODE` as a named constant in `smerds.ts`, not inlined in each mapper
- CPU/mem/uptime/restarts will be filled in once T32 extends the smerd response