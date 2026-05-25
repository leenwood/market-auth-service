.PHONY: build run test lint migrate-up migrate-down generate-keys docker-up docker-down

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

migrate-up:
	psql "$(DATABASE_DSN)" -f migrations/000001_create_users.up.sql

migrate-down:
	psql "$(DATABASE_DSN)" -f migrations/000001_create_users.down.sql

generate-keys:
	openssl genrsa -out private.pem 2048
	openssl rsa -in private.pem -pubout -out public.pem
	@echo "Export keys:"
	@echo "  export JWT_PRIVATE_KEY=\"\$$(cat private.pem)\""
	@echo "  export JWT_PUBLIC_KEY=\"\$$(cat public.pem)\""

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

tidy:
	go mod tidy
