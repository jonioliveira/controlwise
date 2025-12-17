'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/Toast'
import { getErrorMessage } from '@/lib/api-error'
import type { Session, SessionWithDetails, CalendarEvent, SessionStatus, SessionType } from '@/types'

// Query keys factory
export const sessionKeys = {
  all: ['sessions'] as const,
  lists: () => [...sessionKeys.all, 'list'] as const,
  list: (filters: SessionFilters) => [...sessionKeys.lists(), filters] as const,
  details: () => [...sessionKeys.all, 'detail'] as const,
  detail: (id: string) => [...sessionKeys.details(), id] as const,
  calendar: (filters: CalendarFilters) => [...sessionKeys.all, 'calendar', filters] as const,
  stats: () => [...sessionKeys.all, 'stats'] as const,
}

interface SessionFilters {
  therapist_id?: string
  patient_id?: string
  status?: SessionStatus
  from_date?: string
  to_date?: string
  page?: number
  limit?: number
}

interface SessionsResponse {
  data: {
    sessions: SessionWithDetails[]
    total: number
    page: number
    limit: number
  }
}

interface SessionResponse {
  data: {
    session: SessionWithDetails
  }
}

interface SessionMutationResponse {
  data: {
    session: Session
  }
}

interface CalendarResponse {
  data: {
    events: CalendarEvent[]
  }
}

interface CalendarFilters {
  therapist_id?: string
  from_date: string
  to_date: string
}

interface SessionStatsResponse {
  data: {
    total_today: number
    total_week: number
    pending: number
    confirmed: number
    completed_today: number
    revenue_today_cents: number
    revenue_week_cents: number
  }
}

export function useSessions(filters: SessionFilters = {}) {
  const params = new URLSearchParams()
  if (filters.therapist_id) params.set('therapist_id', filters.therapist_id)
  if (filters.patient_id) params.set('patient_id', filters.patient_id)
  if (filters.status) params.set('status', filters.status)
  if (filters.from_date) params.set('from_date', filters.from_date)
  if (filters.to_date) params.set('to_date', filters.to_date)
  if (filters.page) params.set('page', String(filters.page))
  if (filters.limit) params.set('limit', String(filters.limit))

  return useQuery({
    queryKey: sessionKeys.list(filters),
    queryFn: async () => {
      const response = await api.get<SessionsResponse>(`/sessions?${params.toString()}`)
      return response.data?.data
    },
  })
}

export function useSession(id: string | null) {
  return useQuery({
    queryKey: sessionKeys.detail(id || ''),
    queryFn: async () => {
      if (!id) return null
      const response = await api.get<SessionResponse>(`/sessions/${id}`)
      return response.data?.data?.session
    },
    enabled: !!id,
  })
}

export function useCalendarEvents(filters: CalendarFilters) {
  const params = new URLSearchParams()
  if (filters.therapist_id) params.set('therapist_id', filters.therapist_id)
  params.set('from_date', filters.from_date)
  params.set('to_date', filters.to_date)

  return useQuery({
    queryKey: sessionKeys.calendar(filters),
    queryFn: async () => {
      const response = await api.get<CalendarResponse>(`/sessions/calendar?${params.toString()}`)
      return response.data?.data?.events
    },
    enabled: !!filters.from_date && !!filters.to_date,
  })
}

export function useSessionStats() {
  return useQuery({
    queryKey: sessionKeys.stats(),
    queryFn: async () => {
      const response = await api.get<SessionStatsResponse>('/sessions/stats')
      return response.data?.data
    },
  })
}

interface CreateSessionInput {
  therapist_id: string
  patient_id: string
  scheduled_at: string
  duration_minutes?: number
  price_cents?: number
  session_type?: SessionType
  notes?: string
}

export function useCreateSession() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (data: CreateSessionInput) => {
      const response = await api.post<SessionMutationResponse>('/sessions', data)
      return response.data?.data?.session
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
      success('Sessão criada', 'A sessão foi agendada com sucesso.')
    },
    onError: (err) => {
      error('Erro ao criar sessão', getErrorMessage(err))
    },
  })
}

interface UpdateSessionInput extends Partial<CreateSessionInput> {
  status?: SessionStatus
}

export function useUpdateSession() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateSessionInput }) => {
      const response = await api.put<SessionMutationResponse>(`/sessions/${id}`, data)
      return response.data?.data?.session
    },
    onSuccess: () => {
      success('Sessão atualizada', 'A sessão foi atualizada com sucesso.')
    },
    onError: (err) => {
      error('Erro ao atualizar sessão', getErrorMessage(err))
    },
    onSettled: (_, __, { id }) => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}

export function useDeleteSession() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/sessions/${id}`)
      return id
    },
    onSuccess: () => {
      success('Sessão eliminada', 'A sessão foi removida com sucesso.')
    },
    onError: (err) => {
      error('Erro ao eliminar sessão', getErrorMessage(err))
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}

export function useConfirmSession() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await api.post<SessionMutationResponse>(`/sessions/${id}/confirm`)
      return response.data?.data?.session
    },
    // Optimistic update
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: sessionKeys.detail(id) })
      const previousSession = queryClient.getQueryData<SessionWithDetails>(sessionKeys.detail(id))

      if (previousSession) {
        queryClient.setQueryData<SessionWithDetails>(sessionKeys.detail(id), {
          ...previousSession,
          status: 'confirmed',
        })
      }

      return { previousSession }
    },
    onError: (err, id, context) => {
      if (context?.previousSession) {
        queryClient.setQueryData(sessionKeys.detail(id), context.previousSession)
      }
      error('Erro ao confirmar sessão', getErrorMessage(err))
    },
    onSuccess: () => {
      success('Sessão confirmada', 'A sessão foi confirmada com sucesso.')
    },
    onSettled: (_, __, id) => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}

export function useCancelSession() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ id, reason }: { id: string; reason?: string }) => {
      const response = await api.post<SessionMutationResponse>(`/sessions/${id}/cancel`, { reason })
      return response.data?.data?.session
    },
    onSuccess: () => {
      success('Sessão cancelada', 'A sessão foi cancelada.')
    },
    onError: (err) => {
      error('Erro ao cancelar sessão', getErrorMessage(err))
    },
    onSettled: (_, __, { id }) => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}

export function useCompleteSession() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ id, notes }: { id: string; notes?: string }) => {
      const response = await api.post<SessionMutationResponse>(`/sessions/${id}/complete`, { notes })
      return response.data?.data?.session
    },
    onSuccess: () => {
      success('Sessão concluída', 'A sessão foi marcada como concluída.')
    },
    onError: (err) => {
      error('Erro ao concluir sessão', getErrorMessage(err))
    },
    onSettled: (_, __, { id }) => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}

export function useMarkNoShow() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (id: string) => {
      const response = await api.post<SessionMutationResponse>(`/sessions/${id}/no-show`)
      return response.data?.data?.session
    },
    onSuccess: () => {
      success('Falta registada', 'A sessão foi marcada como falta.')
    },
    onError: (err) => {
      error('Erro ao registar falta', getErrorMessage(err))
    },
    onSettled: (_, __, id) => {
      queryClient.invalidateQueries({ queryKey: sessionKeys.all })
    },
  })
}
