-- name: ListWhitelistedImages :many
SELECT id, repository, created_at
FROM velez.whitelisted_images
ORDER BY repository;

-- name: InsertWhitelistedImage :one
-- adds the repository; returns the row id only when this call inserted it
INSERT INTO velez.whitelisted_images (repository)
VALUES ($1)
ON CONFLICT (repository) DO NOTHING
RETURNING id;

-- name: DeleteWhitelistedImage :exec
DELETE FROM velez.whitelisted_images
WHERE id = $1;

-- name: DeleteWhitelistedImageByRepository :exec
DELETE FROM velez.whitelisted_images
WHERE repository = $1;
