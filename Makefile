.PHONY: run admin migrate seed build docker-build docker-run docker-admin tidy db-up db-down

run:
	go run ./cmd/server

admin:
	go run ./cmd/server admin

migrate:
	go run ./cmd/server migrate

seed:
	go run ./cmd/server seed

build:
	go build -o bin/useria ./cmd/server

docker-build:
	docker build -t useria .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env useria

docker-admin:
	docker run --rm -p 9080:9080 --env-file .env useria admin

tidy:
	go mod tidy

# Local Postgres for development (docker).
db-up:
	docker run --rm -d --name useria-pg -e POSTGRES_PASSWORD=password -e POSTGRES_DB=saas_oauth2 -p 5432:5432 postgres:16

db-down:
	docker rm -f useria-pg
