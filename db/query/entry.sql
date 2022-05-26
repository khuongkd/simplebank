-- name: CreateEntry :one
INSERT INTO entries(
  account_id,
  amount
) VALUES ($1, $2)
RETURNING *;

-- name: UpdateEntry :one
UPDATE entries
SET amount = $1
WHERE id = $2
RETURNING *;

-- name: DeleteEntry :exec
DELETE FROM entries WHERE id = $1;

-- name: GetEntry :one
SELECT * FROM entries WHERE id = $1;

-- name: ListEntries :many
SELECT * FROM entries ORDER BY id Limit $1 OFFSET $2;