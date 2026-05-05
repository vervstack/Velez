-- name: ListPlugins :many
SELECT velez.plugins.plugin_type,
       velez.plugins.service_id,
       array_agg(depl.status)::text[] AS statuses
FROM velez.plugins AS plugins
         LEFT JOIN velez.deployment_specifications AS dep_spec
                   ON dep_spec.service_id = plugins.service_id
         LEFT JOIN velez.deployments AS depl
                   ON depl.spec_id = dep_spec.id
GROUP BY plugins.plugin_type,
         plugins.service_id;

-- name: UpsertPlugin :exec
INSERT INTO velez.plugins (plugin_type, service_id)
VALUES ($1, $2)
ON CONFLICT (plugin_type)
    DO UPDATE
    SET service_id = EXCLUDED.service_id;
