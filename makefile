postgres:
	docker run --name postgres16 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=123456 -d postgres:16.0-alpine
createdb:
	docker exec -it postgres16 createdb --username=root --owner=root simple_bank
dropdb:
	docker exec -it postgres16 dropdb simple_bank
migrateup:
	migrate -path db/migration -database "postgresql://root:123456@localhost:5432/simple_bank?sslmode=disable" -verbose up
migrateup1:
	migrate -path db/migration -database "postgresql://root:123456@localhost:5432/simple_bank?sslmode=disable" -verbose up 1
migratedown:
	migrate -path db/migration -database "postgresql://root:123456@localhost:5432/simple_bank?sslmode=disable" -verbose down
migratedown1:
	migrate -path db/migration -database "postgresql://root:123456@localhost:5432/simple_bank?sslmode=disable" -verbose down 1
sqlc:
	docker run --rm -v "D:\golang_sdk\go\SimpleBank:/src" -w /src sqlc/sqlc:1.21.0 generate
test:
	go test -v -cover ./...
server:
	go run main.go
mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/GGjahoon/MySimpleBank/db/sqlc Store
proto:
	.\set.bat
evans:
	evans --host localhost --port 9090 -r repl
.PHONY: createdb  dropdb migrateup migratedown sqlc server mock migrateup1 migratedown1 proto test evans