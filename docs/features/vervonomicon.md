# Vervonomicon

The **vervonomicon** is an index file baked into a Docker image that describes the image as a Verv service resource. It
points to a set of descriptor files (deploy config, secrets, config schema, etc.) co-located in the same directory. When
Velez pulls an image that contains a vervonomicon it can automatically register the service, apply its configuration
schema, wire up secrets, and prepare it for deployments — no manual API calls required.

## File location (inside the image)

```
<image root>/
└── .verv/
    └── <service_name>/
        ├── vervonomicon.yaml   ← index
        ├── deploy.yaml
        ├── secrets.yaml
        └── config.schema.json
```

The directory name is the authoritative service name used when registering with the cluster.

## Index file structure

```yaml
# vervonomicon.yaml

# Optional. Vervonomicon format version.
# Defaults to the latest supported version when omitted.
version: "1"

# Human-readable description shown in the Velez UI.
description: "Short description of what this service does"

# Links to external resources.
links:
  docs: "https://wiki.example.com/my-service"
  repo: "https://github.com/example/my-service"

# Arbitrary labels for filtering in the UI.
tags:
  - backend
  - postgres

# Paths to descriptor files, relative to this file.
# All entries are optional — omit any file you don't need.
files:
  deploy: deploy.yaml
  secrets: secrets.yaml
  config_schema: config.schema.json
```

## Descriptor files

### `deploy.yaml`

Defines the default deploy template for the service. Individual deployments can override specific fields.

```yaml
ports:
  - container: 8080
    host: 8080
resources:
  cpu_limit: "0.5"
  memory_limit: "256m"
restart_policy: on-failure
env:
  LOG_LEVEL: info
```

### `secrets.yaml`

Declares which secrets must be injected as environment variables at deploy time. Only references are stored here —
values are never written to the file.

```yaml
- env_key: DB_PASSWORD
  secret_ref: my-service/db-password
- env_key: API_TOKEN
  secret_ref: my-service/api-token
```

### `config.schema.json`

A JSON Schema (draft-07 or later) describing the service's matreshka configuration. Velez uses it to validate config
values and render a typed settings editor in the UI.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["db_host"],
  "properties": {
    "db_host": { "type": "string", "description": "PostgreSQL host" },
    "db_port": { "type": "integer", "default": 5432 },
    "feature_flags": {
      "type": "object",
      "additionalProperties": { "type": "boolean" }
    }
  }
}
```

## How Velez processes the image

1. **Discovery** — when an image is pulled Velez checks for `.verv/` at the image root. If found, it reads the index
   file from `.verv/<service_name>/vervonomicon.yaml`.
2. **Version check** — the `version` field selects the parser. If omitted the latest version is used.
3. **Service registration** — if the service does not exist in the cluster it is created using the metadata from the
   index (`description`, `tags`, `links`). Idempotent — safe to re-run on every pull.
4. **Config schema upload** — if `files.config_schema` is set, the referenced file is stored so the UI can render a
   typed configuration editor.
5. **Secret binding** — if `files.secrets` is set, the declared refs are recorded; the cluster injects them as
   environment variables when a container is launched.
6. **Deploy template** — if `files.deploy` is set, the referenced file becomes the default deploy template for the
   service.

## Automatic service creation flow

```
image pulled by Velez
    │
    ▼
extract .verv/<name>/vervonomicon.yaml
    │
    ▼
POST /api  CreateService        (idempotent)
    │
    ▼
PUT  /api  UploadConfigSchema   (if config.schema.json present)
    │
    ▼
PUT  /api  BindSecrets          (if secrets.yaml present)
    │
    ▼
Service is ready for deployments
```

## Rules and constraints

- The `.verv/` directory name is the authoritative service name — it must be unique across the cluster.
- `version` is optional; when omitted Velez uses the latest supported format version.
- Secret refs point to paths in the cluster secret store; no secret values are ever written to any descriptor file.
- All `files` entries are optional — a minimal vervonomicon with only `version` and `description` is valid.
- All files are safe to commit to the repository and bake into the image; they contain no credentials.
