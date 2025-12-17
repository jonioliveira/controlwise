'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/Toast'
import { getErrorMessage } from '@/lib/api-error'
import type {
  Workflow,
  WorkflowWithStats,
  WorkflowState,
  WorkflowTransition,
  WorkflowTrigger,
  WorkflowAction,
  MessageTemplate,
  WorkflowExecutionLog,
  ScheduledJob,
} from '@/types'

// Query keys factory
export const workflowKeys = {
  all: ['workflows'] as const,
  lists: () => [...workflowKeys.all, 'list'] as const,
  list: (module?: string) => [...workflowKeys.lists(), { module }] as const,
  details: () => [...workflowKeys.all, 'detail'] as const,
  detail: (id: string) => [...workflowKeys.details(), id] as const,
  templates: () => ['templates'] as const,
  templateList: (channel?: string) => [...workflowKeys.templates(), 'list', { channel }] as const,
  template: (id: string) => [...workflowKeys.templates(), id] as const,
  executionLogs: () => ['execution-logs'] as const,
  executionLogList: (filters: ExecutionLogFilters) => [...workflowKeys.executionLogs(), filters] as const,
  scheduledJobs: () => ['scheduled-jobs'] as const,
  scheduledJobList: (status?: string) => [...workflowKeys.scheduledJobs(), { status }] as const,
}

interface WorkflowsResponse {
  data: {
    workflows: WorkflowWithStats[]
  }
}

// ============ Workflow Hooks ============

export function useWorkflows(module?: string) {
  const params = new URLSearchParams()
  if (module) params.set('module', module)

  return useQuery({
    queryKey: workflowKeys.list(module),
    queryFn: async () => {
      const url = params.toString() ? `/workflows?${params.toString()}` : '/workflows'
      const response = await api.get<WorkflowsResponse>(url)
      return response.data?.data?.workflows ?? []
    },
  })
}

export function useWorkflow(id: string | null) {
  return useQuery({
    queryKey: workflowKeys.detail(id || ''),
    queryFn: async () => {
      if (!id) return null
      const response = await api.get<{ data: Workflow }>(`/workflows/${id}`)
      return response.data?.data
    },
    enabled: !!id,
  })
}

interface CreateWorkflowInput {
  name: string
  description?: string
  module: 'appointments' | 'construction'
  entity_type: 'session' | 'budget' | 'project'
  is_default?: boolean
}

export function useCreateWorkflow() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (data: CreateWorkflowInput) => {
      const response = await api.post<{ data: Workflow }>('/workflows', data)
      return response.data?.data
    },
    onSuccess: (workflow) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() })
      success('Workflow criado', `${workflow.name} foi criado com sucesso.`)
    },
    onError: (err) => {
      error('Erro ao criar workflow', getErrorMessage(err))
    },
  })
}

interface UpdateWorkflowInput {
  name?: string
  description?: string
  is_active?: boolean
  is_default?: boolean
}

export function useUpdateWorkflow() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateWorkflowInput }) => {
      const response = await api.put<{ data: Workflow }>(`/workflows/${id}`, data)
      return response.data?.data
    },
    onSuccess: (workflow) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() })
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflow.id) })
      success('Workflow atualizado', `${workflow.name} foi atualizado com sucesso.`)
    },
    onError: (err) => {
      error('Erro ao atualizar workflow', getErrorMessage(err))
    },
  })
}

export function useDeleteWorkflow() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/workflows/${id}`)
      return id
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() })
      success('Workflow eliminado', 'O workflow foi removido com sucesso.')
    },
    onError: (err) => {
      error('Erro ao eliminar workflow', getErrorMessage(err))
    },
  })
}

export function useDuplicateWorkflow() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ id, name }: { id: string; name: string }) => {
      const response = await api.post<{ data: Workflow }>(`/workflows/${id}/duplicate`, { name })
      return response.data?.data
    },
    onSuccess: (workflow) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() })
      success('Workflow duplicado', `${workflow.name} foi criado com sucesso.`)
    },
    onError: (err) => {
      error('Erro ao duplicar workflow', getErrorMessage(err))
    },
  })
}

// ============ State Hooks ============

interface CreateStateInput {
  name: string
  display_name: string
  description?: string
  state_type: 'initial' | 'intermediate' | 'final'
  color?: string
  position: number
}

export function useCreateState() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ workflowId, data }: { workflowId: string; data: CreateStateInput }) => {
      const response = await api.post<{ data: WorkflowState }>(`/workflows/${workflowId}/states`, data)
      return response.data?.data
    },
    onSuccess: (state, { workflowId }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Estado criado', `${state.display_name} foi adicionado.`)
    },
    onError: (err) => {
      error('Erro ao criar estado', getErrorMessage(err))
    },
  })
}

export function useUpdateState() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ workflowId, stateId, data }: { workflowId: string; stateId: string; data: CreateStateInput }) => {
      const response = await api.put<{ data: WorkflowState }>(`/workflows/${workflowId}/states/${stateId}`, data)
      return response.data?.data
    },
    onSuccess: (state, { workflowId }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Estado atualizado', `${state.display_name} foi atualizado.`)
    },
    onError: (err) => {
      error('Erro ao atualizar estado', getErrorMessage(err))
    },
  })
}

export function useDeleteState() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ workflowId, stateId }: { workflowId: string; stateId: string }) => {
      await api.delete(`/workflows/${workflowId}/states/${stateId}`)
      return stateId
    },
    onSuccess: (_, { workflowId }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Estado eliminado', 'O estado foi removido.')
    },
    onError: (err) => {
      error('Erro ao eliminar estado', getErrorMessage(err))
    },
  })
}

export function useReorderStates() {
  const queryClient = useQueryClient()
  const { error } = useToast()

  return useMutation({
    mutationFn: async ({ workflowId, stateIds }: { workflowId: string; stateIds: string[] }) => {
      await api.put(`/workflows/${workflowId}/states/reorder`, { state_ids: stateIds })
    },
    onSuccess: (_, { workflowId }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
    },
    onError: (err) => {
      error('Erro ao reordenar estados', getErrorMessage(err))
    },
  })
}

// ============ Trigger Hooks ============

interface CreateTriggerInput {
  state_id?: string
  transition_id?: string
  trigger_type: 'on_enter' | 'on_exit' | 'time_before' | 'time_after' | 'recurring'
  time_offset_minutes?: number
  time_field?: string
  recurring_cron?: string
  conditions?: Record<string, unknown>
}

export function useCreateTrigger() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ workflowId, data }: { workflowId: string; data: CreateTriggerInput }) => {
      const response = await api.post<{ data: WorkflowTrigger }>(`/workflows/${workflowId}/triggers`, data)
      return response.data?.data
    },
    onSuccess: (_, { workflowId }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Gatilho criado', 'O gatilho foi adicionado.')
    },
    onError: (err) => {
      error('Erro ao criar gatilho', getErrorMessage(err))
    },
  })
}

export function useUpdateTrigger() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ triggerId, workflowId, data }: { triggerId: string; workflowId: string; data: Partial<CreateTriggerInput> & { is_active?: boolean } }) => {
      const response = await api.put<{ data: WorkflowTrigger }>(`/triggers/${triggerId}`, data)
      return { trigger: response.data?.data, workflowId }
    },
    onSuccess: ({ workflowId }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Gatilho atualizado', 'O gatilho foi atualizado.')
    },
    onError: (err) => {
      error('Erro ao atualizar gatilho', getErrorMessage(err))
    },
  })
}

export function useDeleteTrigger() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ triggerId, workflowId }: { triggerId: string; workflowId: string }) => {
      await api.delete(`/triggers/${triggerId}`)
      return workflowId
    },
    onSuccess: (workflowId) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Gatilho eliminado', 'O gatilho foi removido.')
    },
    onError: (err) => {
      error('Erro ao eliminar gatilho', getErrorMessage(err))
    },
  })
}

// ============ Action Hooks ============

interface CreateActionInput {
  action_type: 'send_whatsapp' | 'send_email' | 'update_field' | 'create_task'
  action_order?: number
  template_id?: string
  action_config?: Record<string, unknown>
}

export function useCreateAction() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ triggerId, workflowId, data }: { triggerId: string; workflowId: string; data: CreateActionInput }) => {
      const response = await api.post<{ data: WorkflowAction }>(`/triggers/${triggerId}/actions`, data)
      return { action: response.data?.data, workflowId }
    },
    onSuccess: ({ workflowId }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Ação criada', 'A ação foi adicionada.')
    },
    onError: (err) => {
      error('Erro ao criar ação', getErrorMessage(err))
    },
  })
}

export function useUpdateAction() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ actionId, workflowId, data }: { actionId: string; workflowId: string; data: Partial<CreateActionInput> & { is_active?: boolean } }) => {
      const response = await api.put<{ data: WorkflowAction }>(`/actions/${actionId}`, data)
      return { action: response.data?.data, workflowId }
    },
    onSuccess: ({ workflowId }) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Ação atualizada', 'A ação foi atualizada.')
    },
    onError: (err) => {
      error('Erro ao atualizar ação', getErrorMessage(err))
    },
  })
}

export function useDeleteAction() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ actionId, workflowId }: { actionId: string; workflowId: string }) => {
      await api.delete(`/actions/${actionId}`)
      return workflowId
    },
    onSuccess: (workflowId) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.detail(workflowId) })
      success('Ação eliminada', 'A ação foi removida.')
    },
    onError: (err) => {
      error('Erro ao eliminar ação', getErrorMessage(err))
    },
  })
}

// ============ Template Hooks ============

interface TemplatesResponse {
  data: {
    templates: MessageTemplate[]
  }
}

export function useTemplates(channel?: string) {
  const params = new URLSearchParams()
  if (channel) params.set('channel', channel)

  return useQuery({
    queryKey: workflowKeys.templateList(channel),
    queryFn: async () => {
      const url = params.toString() ? `/templates?${params.toString()}` : '/templates'
      const response = await api.get<TemplatesResponse>(url)
      return response.data?.data?.templates ?? []
    },
  })
}

export function useTemplate(id: string | null) {
  return useQuery({
    queryKey: workflowKeys.template(id || ''),
    queryFn: async () => {
      if (!id) return null
      const response = await api.get<{ data: MessageTemplate }>(`/templates/${id}`)
      return response.data?.data
    },
    enabled: !!id,
  })
}

interface CreateTemplateInput {
  name: string
  channel: 'whatsapp' | 'email'
  subject?: string
  body: string
  variables?: { name: string; description: string }[]
}

export function useCreateTemplate() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (data: CreateTemplateInput) => {
      const response = await api.post<{ data: MessageTemplate }>('/templates', data)
      return response.data?.data
    },
    onSuccess: (template) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.templates() })
      success('Modelo criado', `${template.name} foi criado com sucesso.`)
    },
    onError: (err) => {
      error('Erro ao criar modelo', getErrorMessage(err))
    },
  })
}

export function useUpdateTemplate() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Partial<CreateTemplateInput> & { is_active?: boolean } }) => {
      const response = await api.put<{ data: MessageTemplate }>(`/templates/${id}`, data)
      return response.data?.data
    },
    onSuccess: (template) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.templates() })
      queryClient.invalidateQueries({ queryKey: workflowKeys.template(template.id) })
      success('Modelo atualizado', `${template.name} foi atualizado com sucesso.`)
    },
    onError: (err) => {
      error('Erro ao atualizar modelo', getErrorMessage(err))
    },
  })
}

export function useDeleteTemplate() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/templates/${id}`)
      return id
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.templates() })
      success('Modelo eliminado', 'O modelo foi removido com sucesso.')
    },
    onError: (err) => {
      error('Erro ao eliminar modelo', getErrorMessage(err))
    },
  })
}

// ============ Default Workflow Hooks ============

interface InitDefaultWorkflowsResponse {
  data: {
    workflows: Workflow[]
  }
  message: string
}

export function useInitDefaultWorkflows() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (module?: 'construction') => {
      const params = module ? `?module=${module}` : ''
      const response = await api.post<InitDefaultWorkflowsResponse>(`/workflows/init-defaults${params}`)
      return response.data?.data?.workflows ?? []
    },
    onSuccess: (workflows) => {
      queryClient.invalidateQueries({ queryKey: workflowKeys.lists() })
      if (workflows.length > 0) {
        success('Workflows criados', `${workflows.length} workflow(s) padrão criado(s) com sucesso.`)
      } else {
        success('Workflows já existem', 'Os workflows padrão já estavam configurados.')
      }
    },
    onError: (err) => {
      error('Erro ao criar workflows', getErrorMessage(err))
    },
  })
}

// ============ Execution Log Hooks ============

export interface ExecutionLogFilters {
  workflow_id?: string
  entity_type?: string
  entity_id?: string
  event_type?: string
  limit?: number
  offset?: number
}

interface ExecutionLogsResponse {
  data: {
    logs: WorkflowExecutionLog[]
    total: number
  }
}

export function useExecutionLogs(filters: ExecutionLogFilters = {}) {
  const params = new URLSearchParams()
  if (filters.workflow_id) params.set('workflow_id', filters.workflow_id)
  if (filters.entity_type) params.set('entity_type', filters.entity_type)
  if (filters.entity_id) params.set('entity_id', filters.entity_id)
  if (filters.event_type) params.set('event_type', filters.event_type)
  if (filters.limit) params.set('limit', String(filters.limit))
  if (filters.offset) params.set('offset', String(filters.offset))

  return useQuery({
    queryKey: workflowKeys.executionLogList(filters),
    queryFn: async () => {
      const url = params.toString() ? `/execution-logs?${params.toString()}` : '/execution-logs'
      const response = await api.get<ExecutionLogsResponse>(url)
      return {
        logs: response.data?.data?.logs ?? [],
        total: response.data?.data?.total ?? 0,
      }
    },
  })
}

// ============ Scheduled Jobs Hooks ============

interface ScheduledJobsResponse {
  data: {
    jobs: ScheduledJob[]
  }
}

export function useScheduledJobs(status?: string, limit?: number) {
  const params = new URLSearchParams()
  if (status) params.set('status', status)
  if (limit) params.set('limit', String(limit))

  return useQuery({
    queryKey: workflowKeys.scheduledJobList(status),
    queryFn: async () => {
      const url = params.toString() ? `/scheduled-jobs?${params.toString()}` : '/scheduled-jobs'
      const response = await api.get<ScheduledJobsResponse>(url)
      return response.data?.data?.jobs ?? []
    },
  })
}

// ============ Workflow Testing Hooks ============

export interface ActionTestResult {
  action: WorkflowAction
  action_type: string
  template?: MessageTemplate
  rendered_subject?: string
  rendered_body: string
  recipient?: string
}

export interface WorkflowTestResult {
  trigger: WorkflowTrigger
  state?: WorkflowState
  transition?: WorkflowTransition
  actions: ActionTestResult[]
  sample_data: Record<string, string>
}

interface TestTriggerResponse {
  data: {
    result: WorkflowTestResult
  }
  message: string
}

export function useTestTrigger() {
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ workflowId, triggerId }: { workflowId: string; triggerId: string }) => {
      const response = await api.post<TestTriggerResponse>(
        `/workflows/${workflowId}/triggers/${triggerId}/test`
      )
      return response.data?.data?.result
    },
    onSuccess: () => {
      success('Teste executado', 'O gatilho foi testado com dados de exemplo.')
    },
    onError: (err) => {
      error('Erro ao testar gatilho', getErrorMessage(err))
    },
  })
}

export interface TemplateVariable {
  name: string
  description: string
  sample_value: string
}

interface VariablesResponse {
  data: {
    entity_type: string
    variables: TemplateVariable[]
  }
}

export function useAvailableVariables(entityType: string) {
  return useQuery({
    queryKey: ['workflow-variables', entityType],
    queryFn: async () => {
      const response = await api.get<VariablesResponse>(`/workflows/variables?entity_type=${entityType}`)
      return response.data?.data?.variables ?? []
    },
    enabled: !!entityType,
  })
}
