-- name: CreateDNSLog :one
INSERT INTO dns_log (dns_query_record,
                     type,
                     ip_address,
                     location)
VALUES ($1, $2, $3, $4) RETURNING *;