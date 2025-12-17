# Guia de Desenvolvimento - ControleWise

## 🎉 O Que Foi Implementado

### Backend - Lógica de Negócio Completa

#### ✅ ClientService (COMPLETO)
**Localização**: `backend/internal/services/client.go`

**Funcionalidades**:
- ✅ `List()` - Listar clientes com paginação
- ✅ `Search()` - Pesquisar por nome ou email
- ✅ `GetByID()` - Obter cliente específico
- ✅ `Create()` - Criar cliente com validações
- ✅ `Update()` - Atualizar cliente
- ✅ `Delete()` - Soft delete com proteção
- ✅ `GetStats()` - Estatísticas para dashboard

**Validações Implementadas**:
- Email único por organização
- Campos obrigatórios (name, email, phone)
- Proteção contra eliminação se tiver worksheets
- Verificação de organização em todas as operações

#### ✅ WorksheetService (COMPLETO)
**Localização**: `backend/internal/services/worksheet.go`

**Funcionalidades**:
- ✅ `List()` - Listar com items e fotos
- ✅ `GetByID()` - Obter worksheet completo
- ✅ `Create()` - Criar com items
- ✅ `Update()` - Atualizar worksheet e items
- ✅ `Review()` - Mudar status (review/approve)
- ✅ `Delete()` - Soft delete com proteção
- ✅ Helper: `getItems()` - Carregar items
- ✅ Helper: `getPhotos()` - Carregar fotos

**Validações Implementadas**:
- Cliente deve existir e pertencer à organização
- Worksheets aprovados não podem ser editados
- Proteção contra eliminação se tiver budgets
- Transações para criar worksheet + items atomicamente

### Frontend - Páginas de Gestão

#### ✅ Página de Clientes (COMPLETA)
**Localização**: `frontend/src/app/dashboard/clients/page.tsx`

**Funcionalidades**:
- ✅ Grid de clientes com cards
- ✅ Pesquisa em tempo real
- ✅ Modal de criação/edição
- ✅ Confirmação de eliminação
- ✅ Validação de formulários
- ✅ Loading states
- ✅ Error handling

#### ✅ Página de Worksheets (COMPLETA)
**Localização**: `frontend/src/app/dashboard/worksheets/page.tsx`

**Funcionalidades**:
- ✅ Listagem de worksheets
- ✅ Status badges
- ✅ Modal de criação/edição
- ✅ Gestão dinâmica de items
- ✅ Seleção de cliente
- ✅ Validações

### Componentes Reutilizáveis Criados

#### ✅ Table Component
**Localização**: `frontend/src/components/ui/Table.tsx`

**Features**:
- Genérico (aceita qualquer tipo de dados)
- Paginação integrada
- Loading state
- Empty state
- Responsivo

**Uso**:
```typescript
<Table
  data={items}
  columns={[
    { header: 'Nome', accessor: 'name' },
    { header: 'Email', accessor: (item) => <a href={item.email}>{item.email}</a> },
  ]}
  keyExtractor={(item) => item.id}
  isLoading={isLoading}
/>
```

#### ✅ Modal Component
**Localização**: `frontend/src/components/ui/Modal.tsx`

**Features**:
- Responsivo
- Fecha com ESC
- 4 tamanhos (sm, md, lg, xl)
- Scroll automático
- Backdrop

**Uso**:
```typescript
<Modal
  isOpen={isOpen}
  onClose={handleClose}
  title="Título"
  size="lg"
>
  <form>...</form>
</Modal>
```

#### ✅ StatusBadge Component
**Localização**: `frontend/src/components/ui/StatusBadge.tsx`

**Features**:
- Suporta todos os status do sistema
- Cores automáticas por tipo
- PriorityBadge incluído

**Uso**:
```typescript
<StatusBadge status="in_progress" />
<PriorityBadge priority="high" />
```

## 📝 Como Implementar as Páginas Restantes

### 1. Budgets (Orçamentos)

#### Backend - BudgetService

```go
// backend/internal/services/budget.go

func (s *BudgetService) Create(ctx context.Context, budget *models.Budget, items []*models.BudgetItem) error {
    // 1. Validar worksheet existe e está aprovado
    // 2. Gerar budget_number (ORG-YYYY-NNNN)
    // 3. Calcular totais (subtotal, tax, total)
    // 4. Criar budget + items em transação
    // 5. Retornar budget criado
}

func (s *BudgetService) Send(ctx context.Context, id, orgID uuid.UUID) error {
    // 1. Verificar status é draft
    // 2. Mudar status para sent
    // 3. Enviar email ao cliente
    // 4. Criar notificação
}

func (s *BudgetService) Approve(ctx context.Context, id, orgID, approverID uuid.UUID) error {
    // 1. Verificar status é sent
    // 2. Mudar status para approved
    // 3. Criar projeto automaticamente
    // 4. Notificar manager
}
```

#### Frontend - Budgets Page

**Copiar estrutura de**: `dashboard/worksheets/page.tsx`

**Adaptar**:
1. Substituir WorkSheet por Budget
2. Adicionar campos de valores (subtotal, tax, total)
3. Adicionar botões: "Enviar", "Aprovar", "Gerar PDF"
4. Mostrar status com cores diferentes
5. Link para worksheet origem

### 2. Projects (Projetos)

#### Backend - ProjectService

```go
func (s *ProjectService) Create(ctx context.Context, project *models.Project) error {
    // Geralmente criado automaticamente quando budget é aprovado
    // 1. Gerar project_number
    // 2. Copiar dados do budget
    // 3. Inicializar progress = 0
}

func (s *ProjectService) UpdateProgress(ctx context.Context, id, orgID uuid.UUID, progress int) error {
    // 1. Validar progress 0-100
    // 2. Atualizar
    // 3. Se progress = 100, mudar status para completed
}

func (s *ProjectService) AddTask(ctx context.Context, projectID uuid.UUID, task *models.Task) error {
    // Criar tarefa associada ao projeto
}
```

#### Frontend - Projects Page

**Estrutura**:
- Card com barra de progresso visual
- Botão para atualizar progresso
- Lista de tarefas no projeto
- Upload de fotos
- Link para budget origem

### 3. Tasks (Tarefas)

#### Backend - TaskService

```go
func (s *TaskService) Assign(ctx context.Context, taskID, userID uuid.UUID) error {
    // 1. Atribuir tarefa
    // 2. Enviar notificação ao utilizador
}

func (s *TaskService) UpdateStatus(ctx context.Context, taskID uuid.UUID, status models.TaskStatus) error {
    // 1. Atualizar status
    // 2. Se completed, preencher completed_at
    // 3. Atualizar progresso do projeto
}
```

#### Frontend - Tasks Page

**Features**:
- Kanban board (Todo, In Progress, Completed)
- Filtro por projeto
- Atribuição de utilizadores
- Due dates com destaque se atrasadas
- Drag & drop (opcional)

### 4. Payments (Pagamentos)

#### Backend - PaymentService

```go
func (s *PaymentService) Create(ctx context.Context, payment *models.Payment) error {
    // 1. Verificar projeto existe
    // 2. Validar valor
    // 3. Criar pagamento
}

func (s *PaymentService) MarkAsPaid(ctx context.Context, id uuid.UUID, paidDate time.Time, method string) error {
    // 1. Atualizar status para paid
    // 2. Preencher paid_at, method
    // 3. Notificar accountant
}

func (s *PaymentService) CheckOverdue(ctx context.Context, orgID uuid.UUID) error {
    // Background job que corre diariamente
    // 1. Encontrar pagamentos pending com due_date < hoje
    // 2. Mudar status para overdue
    // 3. Enviar notificações
}
```

#### Frontend - Payments Page

**Estrutura**:
- Tabela com filtros (pending, paid, overdue)
- Totais no topo
- Botão "Marcar como Pago"
- Indicador visual de atrasos
- Export para Excel (opcional)

## 🔧 Template Rápido para Nova Página

### 1. Service (Backend)

```go
package services

import (
    "context"
    "github.com/controlewise/backend/internal/database"
    "github.com/controlewise/backend/internal/models"
    "github.com/google/uuid"
)

type EntityService struct {
    db *database.DB
}

func NewEntityService(db *database.DB) *EntityService {
    return &EntityService{db: db}
}

func (s *EntityService) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.Entity, int, error) {
    // TODO: Implementar
    return nil, 0, nil
}

func (s *EntityService) GetByID(ctx context.Context, id, orgID uuid.UUID) (*models.Entity, error) {
    // TODO: Implementar
    return nil, nil
}

func (s *EntityService) Create(ctx context.Context, entity *models.Entity) error {
    // TODO: Implementar
    return nil
}

func (s *EntityService) Update(ctx context.Context, id, orgID uuid.UUID, entity *models.Entity) error {
    // TODO: Implementar
    return nil
}

func (s *EntityService) Delete(ctx context.Context, id, orgID uuid.UUID) error {
    // TODO: Implementar
    return nil
}
```

### 2. Handler (Backend)

Copiar `backend/internal/handlers/client.go` e adaptar os tipos.

### 3. Page (Frontend)

```typescript
'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { api } from '@/lib/api'
import { Modal } from '@/components/ui/Modal'

export default function EntitiesPage() {
  const queryClient = useQueryClient()
  const [isModalOpen, setIsModalOpen] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['entities'],
    queryFn: () => api.getEntities(),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteEntity(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['entities'] })
    },
  })

  return (
    <div>
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Entities</h1>
        <button onClick={() => setIsModalOpen(true)} className="btn btn-primary">
          <Plus className="h-5 w-5 mr-2" />
          Nova Entity
        </button>
      </div>

      {/* List */}
      <div className="grid grid-cols-1 gap-6">
        {/* TODO: Render items */}
      </div>

      {/* Modal */}
      {isModalOpen && (
        <EntityFormModal
          onClose={() => setIsModalOpen(false)}
          onSuccess={() => {
            queryClient.invalidateQueries({ queryKey: ['entities'] })
            setIsModalOpen(false)
          }}
        />
      )}
    </div>
  )
}

function EntityFormModal({ onClose, onSuccess }) {
  // TODO: Implementar formulário
  return (
    <Modal isOpen={true} onClose={onClose} title="Nova Entity">
      <form>{/* TODO: Form fields */}</form>
    </Modal>
  )
}
```

## 🚀 Próximos Passos Recomendados

### Prioridade Alta
1. ✅ Implementar **BudgetService** completo
2. ✅ Página de **Budgets** no frontend
3. ✅ Implementar **ProjectService**
4. ✅ Página de **Projects** no frontend

### Prioridade Média
5. ✅ **TaskService** e página de tarefas
6. ✅ **PaymentService** e página de pagamentos
7. ✅ Sistema de upload de fotos funcionando
8. ✅ Relatórios no dashboard

### Prioridade Baixa
9. ✅ Background jobs (pagamentos atrasados, etc)
10. ✅ Export para PDF/Excel
11. ✅ Notificações em tempo real
12. ✅ Testes unitários

## 📚 Recursos Úteis

### Padrões a Seguir
- **Service**: Sempre validar orgID
- **Handler**: Sempre usar middleware.GetOrganizationID()
- **Frontend**: Sempre usar React Query para cache
- **Formulários**: Sempre validar antes de submit

### Comandos Úteis

```bash
# Backend
cd backend
make run              # Executar
go test ./...         # Testar

# Frontend
cd frontend
npm run dev           # Desenvolvimento
npm run build         # Build
npm run lint          # Lint
```

### Debug Tips

1. **Backend não responde**: Ver logs no terminal
2. **CORS errors**: Verificar FRONTEND_URL no .env
3. **DB errors**: Verificar se migrations foram executadas
4. **Auth errors**: Verificar token no localStorage

## 💡 Boas Práticas

### Backend
- ✅ Sempre usar transações para operações múltiplas
- ✅ Soft deletes em vez de hard deletes
- ✅ Validar inputs antes de queries
- ✅ Usar prepared statements (já feito com pgx)
- ✅ Log de erros mas não expor detalhes ao cliente

### Frontend
- ✅ Usar React Query para todas as chamadas API
- ✅ Loading states em todas as operações
- ✅ Error handling com mensagens claras
- ✅ Confirmar ações destrutivas (delete)
- ✅ Validar formulários antes de submeter

## 🎯 Objetivos

Seguindo este guia, em 2-3 dias deves conseguir:
- ✅ Todas as páginas de gestão funcionais
- ✅ CRUD completo para todas as entidades
- ✅ Workflow completo: Cliente → Worksheet → Budget → Project → Payment
- ✅ Sistema pronto para produção (com ajustes de config)

Boa sorte! 🚀
