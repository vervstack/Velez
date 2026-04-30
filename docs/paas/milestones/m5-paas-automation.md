# M5 — PaaS Automation

**Status:** Backlog — flesh out after M4 ships.

## Scope (draft)

- Auto-rollback: if a deploy fails a health check within N minutes, revert to the previous version automatically
- Health-gated promotion: hold traffic on the old version until the new one passes checks
- Scaling policies: min/max replica count, CPU-based autoscale rules
- Scheduled deployments: deploy at a specific time (e.g. maintenance window)
- Webhook triggers: accept a push from a CI system and kick off a deploy automatically
- Secrets management: inject secrets from an external vault rather than plain env vars
