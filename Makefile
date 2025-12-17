# controlwise - Root Makefile
# Container commands use Podman

.PHONY: help up down logs ps restart clean db-shell redis-shell

# Default target
help:
	@echo "controlwise - Available commands:"
	@echo ""
	@echo "Container Management (Podman):"
	@echo "  make up          - Start all services"
	@echo "  make down        - Stop all services"
	@echo "  make restart     - Restart all services"
	@echo "  make logs        - View logs (follow mode)"
	@echo "  make ps          - List running containers"
	@echo "  make clean       - Stop and remove volumes"
	@echo ""
	@echo "Database:"
	@echo "  make db-shell    - Open PostgreSQL shell"
	@echo "  make redis-shell - Open Redis CLI"
	@echo ""
	@echo "Services:"
	@echo "  make up-db       - Start only PostgreSQL and Redis"
	@echo "  make up-all      - Start all services including MinIO and PgAdmin"

# Start core services (PostgreSQL + Redis)
up-db:
	podman compose up -d postgres redis

# Start all services
up up-all:
	podman compose up -d

# Stop all services
down:
	podman compose down

# Restart all services
restart:
	podman compose restart

# View logs
logs:
	podman compose logs -f

# List containers
ps:
	podman compose ps

# Stop and remove volumes (WARNING: destroys data)
clean:
	podman compose down -v

# Open PostgreSQL shell
db-shell:
	podman exec -it controlwise-postgres psql -U controlwise -d controlwise

# Open Redis CLI
redis-shell:
	podman exec -it controlwise-redis redis-cli

# Check service health
health:
	@echo "PostgreSQL:"
	@podman exec controlwise-postgres pg_isready -U controlwise || echo "Not running"
	@echo ""
	@echo "Redis:"
	@podman exec controlwise-redis redis-cli ping || echo "Not running"
