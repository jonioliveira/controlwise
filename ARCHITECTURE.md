# ControleWise - Documentação Técnica

## 📋 Visão Geral

O ControleWise é um sistema de gestão de obras que cobre todo o ciclo desde o contacto inicial do cliente até à conclusão e pagamento do projeto. Desenvolvido com Go no backend e Next.js no frontend, oferece uma solução robusta e escalável para empresas de construção.

## 🏗️ Arquitetura

### Backend (Go 1.25)

**Padrão Arquitetural**: Clean Architecture / Layered Architecture

```
┌─────────────────────────────────────────┐
│           HTTP Handlers                  │
│  (auth, clients, projects, budgets...)   │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│           Services Layer                 │
│    (Business Logic & Orchestration)      │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│         Database Layer                   │
│    (PostgreSQL via pgx/v5)              │
└──────────────────────────────────────────┘
```

**Componentes Principais**:

1. **Handlers**: Recebem requests HTTP, validam input, chamam services
2. **Services**: Contêm a lógica de negócio
3. **Models**: Definem estruturas de dados
4. **Middleware**: Autenticação, CORS, logging
5. **Database**: Conexões e queries

### Frontend (Next.js 16)

**Padrão Arquitetural**: Component-Based Architecture com App Router

```
┌─────────────────────────────────────────┐
│           Pages (App Router)             │
│     /dashboard, /login, /register        │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│          Components Layer                │
│    (Reusable UI Components)              │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│          API Client Layer                │
│    (axios + React Query)                 │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│          Backend API                     │
│         (Go REST API)                    │
└──────────────────────────────────────────┘
```

## 🔐 Sistema de Autenticação

### JWT (JSON Web Tokens)

**Fluxo de Autenticação**:

1. Utilizador faz login com email/password
2. Backend valida credenciais
3. Backend gera JWT token com claims:
   - `user_id`
   - `organization_id`
   - `role`
   - `exp` (expiration)
4. Frontend armazena token em localStorage
5. Requests subsequentes incluem token no header `Authorization: Bearer {token}`

**Middleware de Autenticação**:
- Valida token em cada request
- Extrai informação do utilizador
- Adiciona ao contexto do request

## 🏢 Multi-Tenancy

### Estratégia: Shared Database, Isolated Data

**Implementação**:

1. Cada empresa tem um `organization_id` único
2. Todas as tabelas incluem `organization_id` como foreign key
3. Middleware extrai `organization_id` do token JWT
4. Todas as queries filtram por `organization_id` automaticamente
5. Soft deletes com `deleted_at` para histórico

**Vantagens**:
- Uma base de dados para todas as empresas
- Isolamento de dados garantido
- Escalabilidade horizontal
- Backups simplificados

## 📊 Modelo de Dados

### Entidades Principais

```
Organizations (1) ──< Users (N)
Organizations (1) ──< Clients (N)
Clients (1) ──< WorkSheets (N)
WorkSheets (1) ──< Budgets (N)
Budgets (1) ──< Projects (N)
Projects (1) ──< Tasks (N)
Projects (1) ──< Payments (N)
```

### Workflow de Estados

```
Cliente Contacta
       │
       ▼
┌──────────────┐
│ WorkSheet    │
│ (draft)      │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ WorkSheet    │
│ (review)     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Budget       │
│ (draft)      │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Budget       │
│ (sent)       │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Budget       │
│ (approved)   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Project      │
│ (in_progress)│
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Project      │
│ (completed)  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Payment      │
│ (paid)       │
└──────────────┘
```

## 🔔 Sistema de Notificações

### Tipos de Notificações

1. **In-App**: Armazenadas na base de dados
2. **Email**: Enviadas via SMTP

### Eventos que Geram Notificações

- WorkSheet criada/revista
- Budget enviado/aprovado/rejeitado
- Tarefa atribuída/vencida
- Pagamento pendente/recebido
- Atualização de progresso do projeto

### Implementação

```go
// Service envia notificação
notification := &models.Notification{
    UserID: userID,
    Type: NotificationTypeBudgetApproved,
    Title: "Orçamento Aprovado",
    Message: "O cliente aprovou o orçamento #123",
}

// Cria notificação in-app + email
notificationService.CreateAndEmail(ctx, notification, userEmail)
```

## 📁 Sistema de Ficheiros

### Storage Provider: AWS S3 (ou MinIO localmente)

**Upload Flow**:

1. Frontend envia ficheiro via multipart/form-data
2. Backend valida tipo e tamanho
3. Backend gera nome único: `{org_id}/{uuid}.{ext}`
4. Upload para S3
5. URL guardado na base de dados

**Entidades com Fotos**:
- WorkSheets
- Budgets
- Projects
- Tasks

## 📈 Sistema de Relatórios

### Relatórios Disponíveis

1. **Dashboard**: Estatísticas gerais
   - Clientes ativos
   - Projetos em curso
   - Receita mensal
   - Tarefas pendentes

2. **Projetos**: Análise de projetos
   - Status de projetos
   - Progresso médio
   - Atrasos

3. **Financeiro**: Análise financeira
   - Receita por período
   - Pagamentos pendentes
   - Taxa de aprovação de orçamentos

4. **Clientes**: Análise de clientes
   - Clientes mais ativos
   - Taxa de conversão
   - Valor médio de projeto

## 🔒 Segurança

### Medidas Implementadas

1. **Authentication**: JWT tokens
2. **Authorization**: RBAC (Role-Based Access Control)
3. **Password Security**: bcrypt hashing
4. **SQL Injection**: Parametrized queries
5. **CORS**: Configurado para frontend
6. **Input Validation**: Validação de tipos e formatos
7. **Soft Deletes**: Dados nunca são eliminados permanentemente

### Roles e Permissões

| Role       | Permissões                                      |
|------------|-------------------------------------------------|
| Admin      | Todas                                           |
| Manager    | Criar/editar/aprovar worksheets, budgets, projects |
| Employee   | Ver projetos, completar tarefas                 |
| Client     | Ver os seus próprios projetos e orçamentos      |
| Accountant | Ver/gerir pagamentos                           |

## 🚀 Escalabilidade

### Estratégias de Escala

1. **Horizontal Scaling**: Múltiplas instâncias do backend atrás de load balancer
2. **Database**: PostgreSQL com read replicas
3. **Redis**: Para caching e sessions
4. **S3**: Storage distribuído e escalável
5. **Background Jobs**: Asynq para tarefas assíncronas

### Performance Optimizations

- Connection pooling (PostgreSQL)
- Query indexing
- Redis caching para queries frequentes
- CDN para assets estáticos
- Compressão de imagens

## 📝 Melhorias Futuras

### Short-term (1-3 meses)
- [ ] Implementar todos os endpoints faltantes
- [ ] Testes unitários e integração
- [ ] Validação completa de input
- [ ] Rate limiting
- [ ] API documentation (Swagger)

### Medium-term (3-6 meses)
- [ ] Sistema de relatórios avançado
- [ ] Calendário integrado
- [ ] Chat interno
- [ ] App mobile (React Native)
- [ ] Integração com sistemas de pagamento

### Long-term (6+ meses)
- [ ] IA para estimativas de orçamento
- [ ] Dashboard analítico avançado
- [ ] Integração com ERPs
- [ ] Marketplace de fornecedores
- [ ] API pública para integrações

## 🛠️ Tecnologias Utilizadas

### Backend
- **Language**: Go 1.25
- **Router**: Chi v5
- **Database**: PostgreSQL 16 (via pgx/v5)
- **Cache**: Redis 7
- **Auth**: JWT (golang-jwt)
- **Storage**: AWS S3
- **Email**: SMTP
- **Jobs**: Asynq

### Frontend
- **Framework**: Next.js 16
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State**: React Query + Zustand
- **Forms**: React Hook Form + Zod
- **HTTP**: Axios
- **Icons**: Lucide React

### Infrastructure
- **Database**: PostgreSQL
- **Cache**: Redis
- **Storage**: S3 / MinIO
- **Container**: Docker
- **Deployment**: (TBD)

## 📞 Contacto e Suporte

Para questões técnicas ou suporte, contactar:
- Email: dev@controlewise.io
- Docs: https://docs.controlewise.io

---

**Versão**: 1.0.0  
**Última Atualização**: 2024-11-20
