-- name: CreateVerifyEmail :one
INSERT INTO verify_emails(
                          username,
                          email,
                          secret
)VALUES (
         $1, $2, $3
        )
RETURNING *;

-- name: UpdateVerifyEmail :one
UPDATE verify_emails
SET is_used = true
WHERE
    id = @id
    AND secret = @secret
    AND is_used = false
    AND expire_at > now()
RETURNING *;