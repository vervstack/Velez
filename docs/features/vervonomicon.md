# Vervonomicon

The **vervonomicon** is the declarative deployment descriptor for a service in the Vervstack ecosystem. It lives in a
`.verv/` directory at the root of the service repository and answers one question: *what must the platform create, and
how must it be placed.*

It is deliberately **not** a config file. Application configuration — values, connection strings, tokens — belongs to
**matreshka** (`config/config.yaml`), which Velez already fetches out of the image and fuses into the service's stored
master config. The two never restate the same fact:

| | vervonomicon | matreshka |
|---|---|---|
| Answers | what to create / bind / place | what the app needs to know |
| Contains a `host` / `port` / `pwd` | **never** | always |
| Written by | the developer, in the repo | the developer *and* Velez (resolved values) and operators (UI) |

Declaring a `postgres` resource in the vervonomicon does not describe a connection — it asks Velez to *provision* one.
If the service already has a connection to that resource, nothing is created; see
[Resource reconciliation](#resource-reconciliation).

---

## Directory layout

```
<repo root>/
└── .verv/
    ├── vervonomicon.yaml     ← index + identity (required)
    ├── deployment.yaml       ← the primary container
    ├── resources.yaml        ← what Velez must provision or bind
    ├── ingress.yaml          ← domains, ssl, upstreams
    ├── ingress.conf          ← verbatim webserver config, referenced from ingress.yaml
    ├── auth.yaml             ← access posture
    ├── prod/                 ← per-environment overlays
    │   ├── ingress.yaml
    │   └── ingress.conf
    └── local/
        └── ingress.conf
```

`vervonomicon.yaml` is the only required file. Any other file that is absent is simply ignored.

---

## `vervonomicon.yaml` — index and identity

```yaml
version: "1"

service:
  name: zpotify
  description: Self-hosted spotify alternative
  repo: https://github.com/ruf-dev/zpotify
  tags: [media, backend]

# Default sizing for the app container and for every resource that does not
# override it. See "Boxes" below.
box: small

config:
  # matreshka key paths whose values are masked in the UI and excluded from
  # deployment snapshots. matreshka remains the only place the values live.
  sensitive:
    - telegram_token
    - data_sources.postgres.pwd

# Path overrides. All optional — an omitted field resolves to the conventional
# filename shown, and a missing file is not an error.
deployment: deployment.yaml
resources:  resources.yaml
ingress:    ingress.yaml
auth:       auth.yaml
```

`version` is a major version. A descriptor whose major version the running Velez does not recognise is rejected with a
clear error rather than partially applied.

---

## `deployment.yaml` — the primary container

Every field maps 1:1 onto `CreateSmerd.Request` (see [Mapping](#mapping-onto-createsmerdrequest)).

```yaml
app:
  # image is normally OMITTED. When the descriptor is baked into an image, that
  # image is the subject — naming itself would be circular. Set it only for the
  # repo-url and client-push sources. A deploy request's image always wins.
  # image: ghcr.io/ruf-dev/zpotify:v1.1.2

  command: ""                 # optional; overrides the image entrypoint args
  use_image_ports: false      # publish every port the image EXPOSEs

  ports:
    - port: 80
      protocol: tcp           # tcp | udp
      expose_to: 8080         # optional host port; omit to keep it internal

  volumes:
    - name: zpotify-data
      path: /data

  env:
    SOME_FLAG: "1"

  labels:
    team: media

  healthcheck:
    command: ""               # empty → just wait for the container to reach Running
    interval_second: 5
    timeout_second: 3
    retries: 12

  restart: unless_stopped     # unless_stopped | no | always | on_failure
  restart_failure_count: 5    # on_failure only

  auto_upgrade: true

  box: small                  # optional; overrides the root-level box
  resources:                  # optional; exact values, win over any box
    cpu: 0.5
    ram_mb: 512
    memory_swap_mb: 1024
```

There is no `sidecars` key. Anything that runs as its own container is a **resource** (below). The Verv closed-network
client is attached automatically by Velez and is never declared.

---

## `resources.yaml` — provision or bind

A flat list. Every entry declares its own `type` **and** its own sizing — either a `box` or an exact `resources` block,
which wins when both are present.

```yaml
- name: pg
  type: postgres
  isolation: shared_pool          # shared_pool | separate_instance
  box: medium
  binds_to: data_sources.postgres # matreshka key path the resolved connection is written to

- name: storage
  type: local_volume
  container_path: /data
  sticky: true                    # pins the service to the node holding the volume

- name: media-server
  type: nginx
  box: small
  config:
    - media-server.conf:/etc/nginx/nginx.conf:ro
  volumes:
    - {name: storage, path: /data}
```

Fields by type:

| Field | Applies to | Meaning |
|---|---|---|
| `name` | all | unique within the service; the `service_resources` binding key |
| `type` | all | `postgres`, `redis`, `local_volume`, `nginx`, … |
| `box` / `resources` | all | sizing, exactly as in `deployment.yaml` |
| `isolation` | database types | `shared_pool` (a database in the cluster pool) or `separate_instance` (a dedicated container) |
| `binds_to` | anything with a connection | matreshka key path Velez writes the resolved connection into |
| `container_path`, `sticky` | `local_volume` | mount point inside the app container; whether the service is pinned to the volume's node |
| `config` | container-backed | `<file in .verv>:<path in container>[:<mode>]` |
| `volumes` | container-backed | volumes mounted into the resource's own container |

**A resource entry never carries `host`, `port`, `user` or `pwd`.** Those are resolved by Velez and written into
matreshka.

---

## `ingress.yaml` + config file

Two concerns, deliberately split: `ingress.yaml` declares *what domain must be obtained and how it is terminated*; the
referenced config file is the verbatim webserver config.

```yaml
domains:
  - host: zpotify.ru
    ssl: enabled                 # enabled | disabled | external
    listen: 443
    upstreams:
      api:   {app: true}                 # the primary container
      media: {resource: media-server}    # a container-backed resource
    config: ingress.conf
```

`ingress.conf` is an nginx/angie fragment, mounted verbatim. It is never templated — environment differences are
handled by overlaying a whole replacement file (see below).

---

## `auth.yaml` — access posture

```yaml
forbid_all: true

allow:
  - service: gateway
    paths: ["/api/*"]
```

Recorded and displayed today. ACL rules will be built from it once the access-control milestone lands; nothing is
enforced from this file yet.

---

## Environment overlays

`.verv/<env>/` mirrors the root layout and is merged over it when deploying to environment `<env>`.

```
.verv/ingress.yaml          base
.verv/prod/ingress.yaml     merged over the base when environment == "prod"
.verv/prod/ingress.conf     REPLACES .verv/ingress.conf wholesale
```

Merge rules:

- YAML maps deep-merge by key.
- Lists of objects merge by `name`; an overlay entry with an unseen `name` is appended.
- Lists of scalars, and scalars themselves, are replaced.
- Non-YAML files (`*.conf` and anything else) are replaced wholesale — never merged line by line.
- An overlay directory whose name does not match the deploy's environment is ignored entirely.

---

## Boxes

`box` names a sizing tier. Tiers are stored in `velez.resource_boxes` and seeded with these defaults, which an
administrator may edit or extend with custom tiers:

| Box | cpu | ram_mb | disk_mb |
|---|---|---|---|
| `small` | 0.5 | 512 | 2048 |
| `medium` | 1.0 | 1024 | 8192 |
| `large` | 2.0 | 4096 | 32768 |

Precedence, most specific first: an entry's own `resources` block → an entry's own `box` → the root-level `box` → no
limits at all.

---

## Where the descriptor comes from

Three sources, in decreasing order of robustness. Velez records which one a deployment used.

| Source | Used when | Mechanism |
|---|---|---|
| **Image** | normal deploys; the robust default | `COPY .verv /verv` in the Dockerfile. Velez reads `/verv` out of the image through a throwaway scratch container, the same way it already reads `/app/config/config.yaml`. |
| **Repo URL** | first time a service is added from git | Velez fetches `.verv/` from the repository at the ref being added. |
| **Client push** | a developer deploying to their own environment | The `verv` binary reads the local `.verv/` and sends it with the deploy request. |

Absence of a descriptor is never an error — the service simply deploys the way it does today.

---

## Resource reconciliation

Provisioning a resource and connecting to one are different facts. For each entry in `resources.yaml`, on deploy:

```
does the service already have a connection to this resource?
  ├─ a velez.service_resources binding exists for (service, name), OR
  └─ matreshka data_sources holds an entry for it with a non-placeholder host
        │
        ├─ yes → do NOT create anything. Record status "already_connected".
        │        The service page shows the existing connection and does not
        │        offer to create the resource.
        │
        └─ no  → provision it, then write the resolved connection into the
                 matreshka key named by `binds_to`.
```

A placeholder host is `localhost`, `127.0.0.1`, `0.0.0.0`, or empty — the values a repo's committed `config.yaml`
carries for local development.

---

## How Velez processes the descriptor

```
deploy triggered
    │
    ▼
locate .verv/   (image → repo url → client push)
    │
    ▼
parse vervonomicon.yaml, check version
    │
    ▼
deep-merge the <environment>/ overlay
    │
    ▼
resolve boxes → concrete cpu / ram / disk
    │
    ├─▶ resources: reconcile (skip-if-connected), provision, write matreshka bindings
    ├─▶ deployment: build the primary CreateSmerd.Request
    ├─▶ ingress:    render the config file, apply through the webserver plugin
    └─▶ snapshot the fully resolved descriptor onto the deployment record
```

The snapshot is what makes a deployment reproducible and auditable: `GetVervonomicon` returns both the raw files as
found and the resolved descriptor as applied. Velez does not continuously reconcile — a descriptor takes effect at
deploy time, not before.

---

## Mapping onto `CreateSmerd.Request`

| Descriptor | `CreateSmerd.Request` |
|---|---|
| `service.name` | `name` |
| `service.repo` | `repo` |
| `service.tags` | `labels`, as `verv.tag.<tag>` |
| `app.image` | `image_name` — a deploy request's image always wins |
| `app.command` | `command` |
| `app.ports` | `settings.ports` (`service_port_number`, `protocol`, `exposed_to`) |
| `app.use_image_ports` | `use_image_ports` |
| `app.volumes` | `settings.volumes` (`volume_name`, `container_path`) |
| `app.env` | `env` |
| `app.labels` | `labels` |
| `app.healthcheck` | `healthcheck` |
| `app.restart`, `app.restart_failure_count` | `restart` |
| `app.auto_upgrade` | `auto_upgrade` |
| `app.box` / `app.resources` | `hardware` (`cpu`, `ram_mb`, `memory_swap_mb`) |
| deploy environment | `environment` |
| `ingress` config file | `plain[]` `FileConfig` on the ingress container |
| `resources[]` | provisioning jobs + `velez.service_resources` bindings |

`is_declarative_deploy` is set on every vervonomicon-driven deploy.

---

## Rules and constraints

- `vervonomicon.yaml` is the only required file; everything else is opt-in.
- No descriptor file ever contains a credential, a host or a port of a provisioned resource.
- Every file under `.verv/` is safe to commit and safe to bake into the image.
- The descriptor is a snapshot taken at deploy time, not continuously reconciled desired state.
- A container that runs alongside the app is a **resource**, not a sidecar; there is no sidecar key.
