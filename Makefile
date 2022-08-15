postgres: 
	docker run --name postgres-local -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres:14-alpine

createdb:
	docker exec -it postgres-local createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres-local dropdb simple_bank

# postgres://username:password@localhost:5432/db_name?sslmode=disable
migrateupdb:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" --verbose up

migratedowndb:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" --verbose down

sqlc:
	sqlc generate

logindb:
	docker exec -it postgres-local psql -U root -d simple_bank