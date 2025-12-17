.PHONY: help run build test clean migrate-up migrate-down docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make run          - Run the application"
	@echo "  make build        - Build the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make migrate-up   - Run database migrations up"
	@echo "  make migrate-down - Run database migrations down"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"

run:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

# Database migrations (using golang-migrate)
migrate-up:
	migrate -path migrations -database "postgresql://controlewise:controlewise@localhost:5433/controlewise?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://controlewise:controlewise@localhost:5433/controlewise?sslmode=disable" down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Docker commands
docker-up:
	cd .. && docker-compose up -d

docker-down:
	cd .. && docker-compose down

docker-logs:
	cd .. && docker-compose logs -f

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install tools
tools:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
