APP_NAME := spam-registry

.PHONY: all tidy build run-api run-worker test lint

all: tidy lint test build

# Gestionar dependencias
tidy:
	@echo "🧹 Limpiando módulos..."
	@go mod tidy

# Construir los binarios
build:
	@echo "🏗️ Compilando API y Worker..."
	@go build -o bin/api cmd/api/main.go
	@go build -o bin/worker cmd/worker/main.go

# Ejecutar API localmente
run-api:
	@go run cmd/api/main.go

# Ejecutar Worker localmente
run-worker:
	@go run cmd/worker/main.go

# Calidad de código
lint:
	@echo "🔍 Linting..."
	@golangci-lint run

test:
	@echo "🧪 Testeando..."
	@go test -v ./...