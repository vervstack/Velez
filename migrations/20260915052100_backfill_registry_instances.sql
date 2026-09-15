-- +goose Up
-- +goose StatementBegin

-- Backfills velez.registry_instances for the pre-existing, singleton
-- "registry" plugin (velez.plugins.plugin_type = 'registry') as
-- ContainerRegistryAPI replaces it - see internal/jobs/enable_registry.go
-- and docs/features/pgaas_and_registry_plugin.md section 4 for the flow
-- that originally created these rows. goose runs this migration inside a
-- transaction by default (no "-- +goose NO TRANSACTION" directive), so the
-- insert and the plugins delete below are atomic.
--
-- Column-by-column provenance:
--   service_id  - velez.plugins.service_id (the FK the plugin row already
--                 carries).
--   port        - the registry:2 image's fixed container-internal listening
--                 port (internal/jobs/enable_registry.go's
--                 registryContainerPort constant, 5000). Mirrors
--                 internal/service/service_manager/pgaas/create.go, which
--                 likewise always stores the fixed container port
--                 (pgDefaultPort), never the resolved host-exposed port -
--                 the host-exposed port is assigned by PortManager at
--                 deploy time and isn't duplicated into the satellite
--                 table for either service.
--   username    - velez.registries.username, joined by name (the plugin's
--                 fixed RegistryServiceName "registry" is both
--                 velez.services.name and velez.registries.name for this
--                 singleton) - registerRegistryRowJob wrote it there
--                 verbatim from the same credentials this backfill needs.
--   secret_ref  - velez.registries.secret, which registerRegistryRowJob
--                 already populated with the secret ref string
--                 ("plugin/registry/password"), not a password value.
--
--   ui_port     - OPEN ITEM, no default guess made: the joxit
--                 docker-registry-ui sidecar (registry_ui builtin
--                 descriptor) did not exist before this feature, so no
--                 pre-existing plugin instance has one deployed and no
--                 port for it exists anywhere in Postgres to recover.
--                 Backfilled rows get the sentinel 0 ("no UI sidecar
--                 provisioned yet"); the backend/frontend slices that
--                 consume this table must treat ui_port = 0 as "deploy the
--                 registry_ui sidecar for this instance", not as a real
--                 port. Flagged again in this migration's Down and in the
--                 task 1 handoff report.
INSERT INTO velez.registry_instances (service_id, port, ui_port, username, secret_ref)
SELECT p.service_id,
       5000,
       0,
       r.username,
       r.secret
FROM velez.plugins p
         JOIN velez.services s ON s.id = p.service_id
         JOIN velez.registries r ON r.name = s.name
WHERE p.plugin_type = 'registry'
  AND p.service_id IS NOT NULL
ON CONFLICT (service_id) DO NOTHING;

DELETE FROM velez.plugins WHERE plugin_type = 'registry';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Best-effort reversal only - see 22-data-postgres.md's guidance on
-- migrations that can't be perfectly reversed. Rows this Up created are
-- identified by the ui_port = 0 sentinel (a row created later through the
-- real ContainerRegistryAPI always gets a real ui_port from deploying the
-- registry_ui sidecar, so 0 should only ever match a backfilled row). This
-- diverges from the true pre-Up state if the registry singleton was
-- re-enabled, renamed, or its ui_port genuinely left unset by other means
-- between Up and Down; ui_port itself is not recoverable at all (see Up's
-- comment) and is simply dropped.
INSERT INTO velez.plugins (plugin_type, service_id)
SELECT 'registry', service_id
FROM velez.registry_instances
WHERE ui_port = 0
ON CONFLICT (plugin_type) DO UPDATE SET service_id = EXCLUDED.service_id;

DELETE FROM velez.registry_instances WHERE ui_port = 0;

-- +goose StatementEnd
