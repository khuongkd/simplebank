-- name: CreateTransfer :one
INSERT INTO transfers(from_account_id, to_account_id, amount)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTransfer :one
UPDATE transfers SET amount = $1, from_account_id = $2, to_account_id = $3
WHERE id = $4
RETURNING *;

-- name: DeleteTransfer :exec
DELETE FROM transfers WHERE id = $1;

-- name: GetTransfer :one
SELECT * FROM transfers WHERE id = $1;

-- name: ListTransfers :many
SELECT * FROM transfers ORDER BY id LIMIT $1 OFFSET $2;