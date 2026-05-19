# Vervonomicon

The **vervonomicon** is a declarative deployment descriptor for a service in the Vervstack ecosystem. It lives in a
`.verv/` directory at the root of the service repository — analogous to a Helm chart in Kubernetes, but without the
templating overhead. When Velez pulls an image it looks for a bundled `.verv/` directory; the same directory can also
be used as a first-class source of truth in CI or direct cluster operations.

## Directory layout

```
<repo root>/
└── .verv/
    ├── vervonomicon.yaml     ← index (required)
    ├── webserver.conf        ← nginx config fragment (optional)
    ├── configuration.yaml    ← app config: static values or matreshka references
    ├── resources.yaml        ← hardware limits and cluster placement
    ├── dependencies.yaml     ← external resources (postgres, redis, …)
    ├── network.yaml          ← ports and exposure levels
    ├── auth.yaml             ← inter-service access control
    └── secrets.yaml          ← secret declarations (no values — refs only)
```

All files except `vervonomicon.yaml` are optional. Any file that is absent is simply ignored.

---

## `vervonomicon.yaml` — index

The entry point. Points to all other descriptor files and carries top-level metadata.

```yaml
version: "1"

description: "Short description of what this service does"

links:
  docs: "https://wiki.example.com/my-service"
  repo: "https://github.com/example/my-service"

tags:
  - backend
  - postgres

files:
  webserver: webserver.conf
  configuration: configuration.yaml
  resources: resources.yaml
  dependencies: dependencies.yaml
  network: network.yaml
  auth: auth.yaml
  secrets: secrets.yaml
```

---

## `webserver.conf` — reverse proxy config

An nginx configuration fragment applied to the Vervstack-managed nginx instance (when one is enabled on the cluster).
Standard nginx `server` / `location` block syntax.

```nginx
location /api/my-service/ {
    proxy_pass http://my-service:8080/;
    proxy_set_header Host $host;
}
```

---

## `configuration.yaml` — app configuration

Describes the service's runtime configuration. Can be fully static (values committed to the repo) or dynamic
(values resolved at deploy time from a matreshka path).

```yaml
# Static values committed to the repo
log_level: info
feature_flags:
  new_ui: false

# Dynamic values pulled from matreshka at deploy time
db_host:
  source: matreshka
  path: infra/postgres/host

db_port:
  source: matreshka
  path: infra/postgres/port
  default: 5432
```

---

## `resources.yaml` — hardware and placement

Declares resource limits and preferred cluster placement.

```yaml
resources:
  cpu_limit: "0.5"       # fractional cores
  memory_limit: "256m"
  memory_request: "128m"

placement:
  preferred_clusters:
    - cluster-eu-west
    - cluster-eu-central
  node_labels:
    tier: backend
```

---

## `dependencies.yaml` — external resources

Declares what infrastructure the service needs allocated before it can start. Velez provisions these resources (or
verifies they exist) during the deploy pipeline.

```yaml
postgres:
  - name: my-service-db
    version: "15"
    size: small

redis:
  - name: my-service-cache
    version: "7"

volumes:
  - name: my-service-data
    size: 10Gi
    mount_path: /data
```

---

## `network.yaml` — ports and exposure

Declares which ports the service listens on and how they are exposed. Determines what Velez binds in Docker and what
the reverse proxy can reach.

```yaml
ports:
  - name: grpc
    port: 53890
    expose: vpn-only     # internal | vpn-only | public

  - name: metrics
    port: 9090
    expose: internal
```

---

## `auth.yaml` — inter-service access control

Describes which other services may call this one, and which URI paths they are allowed to reach. Velez uses this
file to build a service dependency graph and enforce network policies.

```yaml
allowed_callers:
  - service: gateway
    paths:
      - /api/*
  - service: billing
    paths:
      - /api/v1/usage

# URIs this service exposes (used for graph rendering in the UI)
endpoints:
  - path: /api/v1/usage
    method: GET
  - path: /api/v1/report
    method: POST
```

---

## `secrets.yaml` — secret declarations

Declares which secrets must be injected as environment variables at deploy time. **Only references are stored here —
values are never written to this file.** Velez reads this file and generates input fields in the cluster secrets
interface, where operators fill in the actual values.

```yaml
- env_key: DB_PASSWORD
  secret_ref: my-service/db-password
  description: "PostgreSQL password for the service user"

- env_key: API_TOKEN
  secret_ref: my-service/api-token
  description: "Upstream API access token"
```

Rules:

- `env_key` is the environment variable name injected into the container.
- `secret_ref` is the path in the cluster secret store.
- `description` is shown as a label in the secrets UI — keep it human-readable.
- This file is safe to commit; it contains no credential values.

---

## How Velez processes the descriptor

```
image pulled / deploy triggered
    │
    ▼
locate .verv/ directory
    │
    ▼
parse vervonomicon.yaml  →  version check
    │
    ├─▶  CreateService (idempotent, uses metadata + tags)
    ├─▶  ProvisionDependencies  (databases, volumes)
    ├─▶  ApplyNetworkPolicy     (network.yaml + auth.yaml)
    ├─▶  SyncConfiguration      (configuration.yaml → matreshka)
    ├─▶  RegisterSecrets        (secrets.yaml → secrets UI)
    ├─▶  ApplyWebserverConfig   (webserver.conf → nginx reload)
    └─▶  Service is ready for deployments
```

## Rules and constraints

- `vervonomicon.yaml` is the only required file; all others are opt-in.
- The directory name under `.verv/` is the authoritative service name — must be unique across the cluster.
- Secret values are never stored in any descriptor file; `secrets.yaml` holds refs only.
- All files are safe to commit to the repository and bake into the Docker image.
- Static `configuration.yaml` values can be committed; matreshka-sourced values are resolved at runtime.
