postgres:
	docker run --name postgres12 -p 5430:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine


execDockerPG:
	docker exec -it postgres12 psql -U root -d simple_bank

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	 migrate -path db/migration -database "postgresql://root:secret@localhost:5430/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5430/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: execDockerPG createdb dropdb postgres migrateup migratedown

# lsof -i :5432
# kill <PID> 
# bind to another port 500?:5234