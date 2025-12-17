# ControleWise - Setup Guide

## 🚀 Quick Start

### Pré-requisitos

- Go 1.25+
- Node.js 20+
- pnpm 9+
- PostgreSQL 16
- Redis 7
- Podman & podman-compose (opcional, para serviços locais)

## Backend Setup

### 1. Configurar Base de Dados

**Opção A: Podman (Recomendado)**
```bash
podman compose up -d postgres redis
```

**Opção B: Instalação Local**
```bash
# PostgreSQL
createdb controlewise
createuser controlewise -P

# Redis
# Instalar através do package manager do seu OS
```

### 2. Configurar Backend

```bash
cd backend

# Copiar configuração
cp .env.example .env

# Editar .env com as suas configurações
nano .env

# Instalar dependências
go mod download

# Instalar ferramentas de migração
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Executar migrações
make migrate-up

# Ou manualmente:
migrate -path migrations -database "postgresql://controlewise:controlewise@localhost:5432/controlewise?sslmode=disable" up

# Executar servidor
make run
```

O servidor estará disponível em `http://localhost:8080`

### 3. Verificar Health Check

```bash
curl http://localhost:8080/health
```

## Frontend Setup

### 1. Configurar Frontend

```bash
cd frontend

# Copiar configuração
cp .env.example .env.local

# Instalar dependências
pnpm install

# Executar em modo de desenvolvimento
pnpm run dev
```

O frontend estará disponível em `http://localhost:3000`

### 2. Build para Produção

```bash
pnpm run build
pnpm start
```

## Configurações Importantes

### Backend (.env)

```env
# Servidor
PORT=8080
ENV=development

# Base de Dados
DB_HOST=localhost
DB_PORT=5432
DB_USER=controlewise
DB_PASSWORD=controlewise
DB_NAME=controlewise

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT (MUDAR EM PRODUÇÃO!)
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# AWS S3 (Opcional)
AWS_REGION=eu-west-1
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
S3_BUCKET=controlewise-files

# Email (Configurar para notificações)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
```

### Frontend (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=ControleWise
```

## Estrutura do Projeto

### Backend (Go)

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── config/                  # Configuração
│   ├── database/                # Conexões DB
│   ├── handlers/                # HTTP handlers
│   ├── middleware/              # Auth, CORS, etc
│   ├── models/                  # Estruturas de dados
│   ├── router/                  # Routing setup
│   ├── services/                # Business logic
│   └── utils/                   # Helpers
├── migrations/                  # Database migrations
├── .env.example
├── go.mod
└── Makefile
```

### Frontend (Next.js)

```
frontend/
├── src/
│   ├── app/
│   │   ├── dashboard/          # Dashboard pages
│   │   ├── login/              # Login page
│   │   ├── register/           # Register page
│   │   ├── layout.tsx          # Root layout
│   │   ├── page.tsx            # Home page
│   │   └── globals.css         # Global styles
│   ├── components/             # React components
│   ├── lib/                    # Utilities
│   │   └── api.ts              # API client
│   └── types/                  # TypeScript types
├── public/                     # Static files
├── .env.example
├── next.config.js
├── tailwind.config.js
└── package.json
```

## API Endpoints Disponíveis

### Autenticação
- `POST /auth/register` - Registar organização e admin
- `POST /auth/login` - Login
- `GET /auth/me` - Obter utilizador atual
- `POST /auth/logout` - Logout

### Clientes
- `GET /clients` - Listar clientes
- `POST /clients` - Criar cliente
- `GET /clients/:id` - Obter cliente
- `PUT /clients/:id` - Atualizar cliente
- `DELETE /clients/:id` - Eliminar cliente

### Folhas de Obra
- `GET /worksheets` - Listar folhas de obra
- `POST /worksheets` - Criar folha de obra
- `POST /worksheets/:id/review` - Rever folha de obra
- `POST /worksheets/:id/photos` - Upload de fotos

### Orçamentos
- `GET /budgets` - Listar orçamentos
- `POST /budgets` - Criar orçamento
- `POST /budgets/:id/send` - Enviar orçamento
- `POST /budgets/:id/approve` - Aprovar orçamento
- `POST /budgets/:id/reject` - Rejeitar orçamento

### Projetos
- `GET /projects` - Listar projetos
- `POST /projects` - Criar projeto
- `PATCH /projects/:id/progress` - Atualizar progresso
- `PATCH /projects/:id/status` - Atualizar status

### Tarefas
- `GET /tasks` - Listar tarefas
- `POST /tasks` - Criar tarefa
- `PATCH /tasks/:id/status` - Atualizar status

### Pagamentos
- `GET /payments` - Listar pagamentos
- `POST /payments` - Criar pagamento
- `POST /payments/:id/mark-paid` - Marcar como pago

### Notificações
- `GET /notifications` - Listar notificações
- `GET /notifications/unread-count` - Contagem não lidas
- `POST /notifications/:id/read` - Marcar como lida

## Comandos Úteis

### Backend

```bash
# Executar servidor
make run

# Build
make build

# Testes
make test

# Criar nova migração
make migrate-create

# Executar migrações
make migrate-up

# Reverter migrações
make migrate-down

# Ver logs Podman
podman compose logs -f
```

### Frontend

```bash
# Desenvolvimento
pnpm run dev

# Build
pnpm run build

# Produção
pnpm start

# Lint
pnpm run lint

# Type check
pnpm run type-check
```

## Deployment

### Backend

1. Build da aplicação:
```bash
go build -o bin/api cmd/api/main.go
```

2. Executar migrações em produção
3. Configurar variáveis de ambiente
4. Executar aplicação

### Frontend

1. Build:
```bash
pnpm run build
```

2. Deploy para Vercel/Netlify ou servidor próprio

## Multi-Tenancy

O sistema está configurado para multi-tenancy ao nível da organização:

- Cada empresa tem o seu próprio `organization_id`
- Todos os dados são isolados por organização
- Middleware valida acesso aos dados da organização correta

## Segurança

- ✅ JWT authentication
- ✅ Password hashing com bcrypt
- ✅ CORS configurado
- ✅ Rate limiting (TODO)
- ✅ Input validation (TODO)
- ✅ SQL injection protection (via parametrized queries)

## Features Implementadas

### ✅ Completo
- Autenticação e autorização
- Multi-tenancy
- Modelos de dados completos
- Migrations de base de dados
- API REST structure
- Frontend com Next.js 16
- Sistema de notificações
- Upload de ficheiros (S3)
- Email service

### 🚧 Por Implementar
- Lógica de negócio completa nos services
- Validação de input detalhada
- Testes unitários
- Testes de integração
- Sistema de relatórios
- Webhooks
- Background jobs para tarefas agendadas

## Troubleshooting

### Base de dados não conecta

```bash
# Verificar se PostgreSQL está a correr
pg_isready -h localhost -p 5432

# Ver logs Podman
podman compose logs postgres
```

### Backend não inicia

```bash
# Verificar variáveis de ambiente
cat .env

# Verificar portas
lsof -i :8080
```

### Frontend não conecta ao backend

1. Verificar `NEXT_PUBLIC_API_URL` em `.env.local`
2. Verificar CORS no backend
3. Ver console do browser para erros

## Suporte

Para questões e suporte, contactar a equipa de desenvolvimento.

## Licença

Proprietary - Todos os direitos reservados.
