include app.env

postgres: ## run db container
	docker run --name postgres14 --network bank-network -p $(POSTGRES_PORT):$(POSTGRES_PORT) -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -d postgres:14-alpine

createdb: ## create database in db container
	docker exec -it postgres14 createdb --username=$(POSTGRES_USER) --owner=$(POSTGRES_USER) schedule

dropdb: ## drop database in db container
	docker exec -it postgres14 dropdb schedule

migrateup: ## migrate up in db container
	migrate -path db/migration -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(HOST):$(POSTGRES_PORT)/$(DB_NAME)?sslmode=disable" -verbose up

migratedown: ## migrate down in db container
	migrate -path db/migration -database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(HOST):$(POSTGRES_PORT)/$(DB_NAME)?sslmode=disable" -verbose down

.PHONY: ##
	createdb dropdb postgres migrateup migratedown

backup: ## backup prod database
	pg_dump -U $(POSTGRES_USER) -h ${HOST} -p $(POSTGRES_PORT) ${DB_NAME} > backup.sql

sync: ## sync prod database to local
	psql -U $(POSTGRES_USER) -h ${HOST} -p $(POSTGRES_PORT) ${STAGE_DB_NAME} < backup.sql     

help: ## List of all commands
	@grep -E '(^[a-zA-Z_0-9-]+:.*?##.*$$)|(^##)' Makefile \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "${G}%-24s${NC} %s\n", $$1, $$2}' \
	| sed -e 's/\[32m## /[33m/' && printf "\n";
