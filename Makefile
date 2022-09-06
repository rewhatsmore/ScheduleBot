include app.env

postgres:
	docker run --name postgres14 --network bank-network -p $(POSTGRES_PORT):$(POSTGRES_PORT) -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -d postgres:14-alpine
createdb:
	docker exec -it postgres14 createdb --username=$(POSTGRES_USER) --owner=$(POSTGRES_USER) shedule
dropdb:
	docker exec -it postgres14 dropdb shedule
migrateup:
	migrate -path db/migration -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(HOST):$(POSTGRES_PORT)/$(DB_NAME)?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migration -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(HOST):$(POSTGRES_PORT)/$(DB_NAME)?sslmode=disable" -verbose down
PHONY: createdb dropdb postgres migrateup migratedown
