.PHONY: dev test coverage coverage-html generate storage-driver infra-up infra-down infra-stop infra-clean db-up redis-up app-up app-down

# Variáveis
ENVIRONMENT ?= development
GO_TEST_FLAGS ?= -v

dev:
	go run cmd/api/main.go

test:
	@echo "Executando testes em ambiente de: test"
	@export ENVIRONMENT=test && go test $(GO_TEST_FLAGS) ./...

coverage:
	@echo "Gerando relatório de cobertura..."
	@export ENVIRONMENT=test && go test -count=1 -coverpkg=./... -coverprofile=coverage.out $$(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...)
	@echo "\n--- Resumo de Cobertura ---"
	@go tool cover -func=coverage.out
	@echo "\n--- Linhas Não Cobertas ---"
	@go tool cover -func=coverage.out | grep -v "100.0%" || echo "Parabéns! 100% de cobertura atingida."

coverage-html:
	@export ENVIRONMENT=test && go test -coverpkg=./... -coverprofile=coverage.out $$(go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' ./...)
	@go tool cover -html=coverage.out

# Geradores
generate:
	go run tools/generator/crud/main.go $(name)

storage-driver:
	go run tools/generator/storage/main.go $(name)

# Infraestrutura
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
