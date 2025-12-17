import { z } from 'zod'

// Workflow schema
export const workflowSchema = z.object({
  name: z
    .string()
    .min(2, 'Nome deve ter pelo menos 2 caracteres')
    .max(100, 'Nome deve ter no máximo 100 caracteres'),
  description: z.string().max(500, 'Descrição deve ter no máximo 500 caracteres').optional(),
  module: z.enum(['appointments', 'construction'], {
    required_error: 'Módulo é obrigatório',
  }),
  entity_type: z.enum(['session', 'budget', 'project'], {
    required_error: 'Tipo de entidade é obrigatório',
  }),
  is_default: z.boolean().default(false),
})

export type WorkflowFormData = z.infer<typeof workflowSchema>

// State schema
export const stateSchema = z.object({
  name: z
    .string()
    .min(2, 'Nome deve ter pelo menos 2 caracteres')
    .max(50, 'Nome deve ter no máximo 50 caracteres')
    .regex(/^[a-z_]+$/, 'Nome deve conter apenas letras minúsculas e underscores'),
  display_name: z
    .string()
    .min(2, 'Nome de exibição deve ter pelo menos 2 caracteres')
    .max(100, 'Nome de exibição deve ter no máximo 100 caracteres'),
  description: z.string().max(255, 'Descrição deve ter no máximo 255 caracteres').optional(),
  state_type: z.enum(['initial', 'intermediate', 'final'], {
    required_error: 'Tipo de estado é obrigatório',
  }),
  color: z
    .string()
    .regex(/^#[0-9A-Fa-f]{6}$/, 'Cor deve ser um código hex válido')
    .optional(),
  position: z.number().int().min(0, 'Posição deve ser um número positivo'),
})

export type StateFormData = z.infer<typeof stateSchema>

// Trigger schema
export const triggerSchema = z.object({
  state_id: z.string().uuid('ID de estado inválido').optional(),
  transition_id: z.string().uuid('ID de transição inválido').optional(),
  trigger_type: z.enum(['on_enter', 'on_exit', 'time_before', 'time_after', 'recurring'], {
    required_error: 'Tipo de gatilho é obrigatório',
  }),
  time_offset_minutes: z.number().int().optional(),
  time_field: z.string().optional(),
  recurring_cron: z.string().optional(),
}).refine(
  (data) => data.state_id || data.transition_id,
  { message: 'Deve selecionar um estado ou transição' }
)

export type TriggerFormData = z.infer<typeof triggerSchema>

// Action schema
export const actionSchema = z.object({
  action_type: z.enum(['send_whatsapp', 'send_email', 'update_field', 'create_task'], {
    required_error: 'Tipo de ação é obrigatório',
  }),
  action_order: z.number().int().min(0, 'Ordem deve ser um número positivo').default(0),
  template_id: z.string().uuid('ID de modelo inválido').optional(),
  action_config: z.record(z.unknown()).optional(),
})

export type ActionFormData = z.infer<typeof actionSchema>

// Message template schema
export const templateSchema = z.object({
  name: z
    .string()
    .min(2, 'Nome deve ter pelo menos 2 caracteres')
    .max(100, 'Nome deve ter no máximo 100 caracteres'),
  channel: z.enum(['whatsapp', 'email'], {
    required_error: 'Canal é obrigatório',
  }),
  subject: z.string().max(255, 'Assunto deve ter no máximo 255 caracteres').optional(),
  body: z
    .string()
    .min(1, 'Corpo da mensagem é obrigatório')
    .max(4000, 'Corpo deve ter no máximo 4000 caracteres'),
  variables: z.array(z.object({
    name: z.string(),
    description: z.string(),
  })).optional(),
})

export type TemplateFormData = z.infer<typeof templateSchema>

// Trigger type display names
export const triggerTypeLabels: Record<string, string> = {
  on_enter: 'Ao entrar no estado',
  on_exit: 'Ao sair do estado',
  time_before: 'Tempo antes',
  time_after: 'Tempo depois',
  recurring: 'Recorrente',
}

// Action type display names
export const actionTypeLabels: Record<string, string> = {
  send_whatsapp: 'Enviar WhatsApp',
  send_email: 'Enviar Email',
  update_field: 'Atualizar campo',
  create_task: 'Criar tarefa',
}

// State type display names
export const stateTypeLabels: Record<string, string> = {
  initial: 'Inicial',
  intermediate: 'Intermediário',
  final: 'Final',
}

// Default colors for states
export const defaultStateColors: Record<string, string> = {
  initial: '#3B82F6', // Blue
  intermediate: '#F59E0B', // Amber
  final: '#22C55E', // Green
}

// Available template variables by entity type
export const templateVariables: Record<string, { name: string; description: string }[]> = {
  session: [
    { name: 'patient_name', description: 'Nome do paciente' },
    { name: 'patient_phone', description: 'Telefone do paciente' },
    { name: 'patient_email', description: 'Email do paciente' },
    { name: 'therapist_name', description: 'Nome do terapeuta' },
    { name: 'session_date', description: 'Data da sessão' },
    { name: 'session_time', description: 'Hora da sessão' },
    { name: 'session_type', description: 'Tipo de sessão' },
    { name: 'amount', description: 'Valor da sessão' },
    { name: 'confirmation_link', description: 'Link de confirmação' },
  ],
  budget: [
    { name: 'client_name', description: 'Nome do cliente' },
    { name: 'client_email', description: 'Email do cliente' },
    { name: 'project_name', description: 'Nome do projeto' },
    { name: 'budget_number', description: 'Número do orçamento' },
    { name: 'budget_total', description: 'Valor total' },
    { name: 'budget_link', description: 'Link para o orçamento' },
    { name: 'approval_link', description: 'Link de aprovação' },
  ],
  project: [
    { name: 'client_name', description: 'Nome do cliente' },
    { name: 'project_name', description: 'Nome do projeto' },
    { name: 'project_number', description: 'Número do projeto' },
    { name: 'status', description: 'Status do projeto' },
    { name: 'progress', description: 'Progresso (%)' },
  ],
}
