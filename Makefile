postgres:
	docker run --name postgres14 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=reginapost -d postgres:14-alpine
createdb:
	docker exec -it postgres14 createdb --username=root --owner=root shedule
dropdb:
	docker exec -it postgres14 dropdb shedule
migrateup:
	migrate -path db/migration -database "postgresql://root:reginapost@localhost:5432/shedule?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migration -database "postgresql://root:reginapost@localhost:5432/shedule?sslmode=disable" -verbose down
PHONY: createdb dropdb postgres migrateup migratedown
