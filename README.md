# ControleWise.io

Sistema de orçamentação, controlo e gestão de obras com suporte multi-tenancy.

## Stack Tecnológico

### Backend
- Go 1.25
- PostgreSQL 16
- Redis 7
- Chi Router
- sqlc (type-safe SQL)
- JWT Authentication

### Frontend
- Next.js 16 (App Router)
- TypeScript
- Tailwind CSS
- React Query
- Shadcn/ui
- Zod

### Tooling
- pnpm (package manager)
- Podman (containers)

## Estrutura do Projeto

```
controlewise/
├── backend/           # API Go
├── frontend/          # Next.js App
├── compose.yml        # Podman compose file
├── Makefile           # Root commands
└── README.md
```

## Setup Rápido

### Pré-requisitos
- Go 1.25+
- Node.js 20+
- pnpm 9+
- Podman
- PostgreSQL 16

### Backend

```bash
cd backend
cp .env.example .env
go mod download
make migrate-up
make run
```

### Frontend

```bash
cd frontend
cp .env.example .env.local
pnpm install
pnpm run dev
```

### Podman (Serviços Locais)

```bash
# Start database and redis
make up-db

# Or start all services (includes MinIO, PgAdmin)
make up

# View logs
make logs

# Stop services
make down
```

## Features

- ✅ Multi-tenancy (várias empresas)
- ✅ RBAC (Role-Based Access Control)
- ✅ Workflow completo: Cliente → Folha de Obra → Orçamento → Projeto → Pagamento
- ✅ Sistema de calendário e tarefas
- ✅ Upload e gestão de fotos
- ✅ Notificações (email, in-app)
- ✅ Relatórios e analytics
- ✅ API REST documentada

## Workflow

1. **Contacto Inicial** - Cliente entra em contacto
2. **Folha de Obra** - Criação e revisão da folha de obra
3. **Orçamento** - Geração e aprovação do orçamento
4. **Em Obra** - Gestão do projeto em execução
5. **Conclusão** - Finalização e pagamento

## Licença

Proprietary
