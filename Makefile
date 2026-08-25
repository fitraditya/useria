.PHONY: run admin migrate seed build docker-build docker-run docker-admin tidy

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
