# Auto-update

Velez can watch app services and move them to newer image tags on its own. The worker is a
plugin: it runs only when config key `auto_update_is_enabled` is true and checks every
`auto_update_check_period_seconds`. With the worker disabled the plugin shows as `disabled` in
`ListPlugins` and per-service labels are still stored, but nothing is upgraded.

## Per-service state

State lives only in container labels:

- `VELEZ_AUTO_UPGRADE` - `true` makes the worker watch the service; `false` stops it.
- `VELEZ_AUTO_UPGRADE_BUMP_LEVEL` - `major`, `minor` or `patch`; absent means `patch`.

Docker cannot change labels of a running container or remove them, so disabling sets
`VELEZ_AUTO_UPGRADE=false` instead of dropping the label.

## Bump levels

Only strict `X.Y.Z` / `vX.Y.Z` tags are compared; the highest accepted tag above the current one wins.

- `major` - take only major bumps.
- `minor` - take minor and major bumps.
- `patch` - take any newer tag.

A non-semver current tag (for example `latest`) moves to the registry's last listed tag.

## API

`AutoUpdateAPI` (`/api/auto_update/`):

- `ListAutoUpdateServices` - app services of the default environment with `is_enabled` and `bump_level`,
  plus `check_period_seconds` and `is_worker_enabled` from config.
- `SetServiceAutoUpdate` - toggle or change the level. This recreates the service through one
  `upgrade_smerd` task (same image, new labels), so a toggle costs one restart. The response carries
  the task's `entity_id` and `action`.
