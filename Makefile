DB_URL=postgresql://root:c4X1BtFHi60rON@localhost:5432/dnsbin?sslmode=disable

network:
	docker network create dnsbin-network

postgres:
	docker run --name postgres --network dnsbin-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=c4X1BtFHi60rON -d postgres:latest

createdb:
	docker exec -it postgres createdb --username=root --owner=root dnsbin

dropdb:
	docker exec -it postgres dropdb dnsbin

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

.PHONY: network postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration sqlc test server