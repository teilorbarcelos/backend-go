.PHONY: dev test coverage generate storage-driver infra-up infra-down infra-stop infra-clean db-up redis-up app-up app-down

dev:
	go run cmd/api/main.go

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Example: make generate name=Product
generate:
	go run tools/generator/crud/main.go $(name)

# Example: make storage-driver name=s3
storage-driver:
	go run tools/generator/storage/main.go $(name)

infra-up:
	docker compose -f docker-compose.infra.yml up -d

infra-stop:
	docker compose -f docker-compose.infra.yml stop

infra-down:
	docker compose -f docker-compose.infra.yml down

infra-clean:
	docker compose -f docker-compose.infra.yml down -v --rmi all

db-up:
	docker compose -f docker-compose.infra.yml up -d db

redis-up:
	docker compose -f docker-compose.infra.yml up -d redis
