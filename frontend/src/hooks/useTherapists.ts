'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/Toast'
import { getErrorMessage } from '@/lib/api-error'
import type { Therapist, WorkingHours } from '@/types'

// Query keys factory
export const therapistKeys = {
  all: ['therapists'] as const,
  lists: () => [...therapistKeys.all, 'list'] as const,
  list: (filters: TherapistFilters) => [...therapistKeys.lists(), filters] as const,
  details: () => [...therapistKeys.all, 'detail'] as const,
  detail: (id: string) => [...therapistKeys.details(), id] as const,
  stats: () => [...therapistKeys.all, 'stats'] as const,
}

interface TherapistFilters {
  search?: string
  is_active?: boolean
  page?: number
  limit?: number
}

interface TherapistsResponse {
  data: {
    therapists: Therapist[]
    total: number
    page: number
    limit: number
  }
}

// Type for cached data (after unwrapping from API response)
interface TherapistsData {
  therapists: Therapist[]
  total: number
  page: number
  limit: number
}

interface TherapistStatsResponse {
  data: {
    total: number
    active: number
    total_sessions_today: number
    total_sessions_week: number
  }
}

export function useTherapists(filters: TherapistFilters = {}) {
  const params = new URLSearchParams()
  if (filters.search) params.set('search', filters.search)
  if (filters.is_active !== undefined) params.set('is_active', String(filters.is_active))
  if (filters.page) params.set('page', String(filters.page))
  if (filters.limit) params.set('limit', String(filters.limit))

  return useQuery({
    queryKey: therapistKeys.list(filters),
    queryFn: async () => {
      const response = await api.get<TherapistsResponse>(`/therapists?${params.toString()}`)
      return response.data?.data
    },
  })
}

export function useTherapist(id: string | null) {
  return useQuery({
    queryKey: therapistKeys.detail(id || ''),
    queryFn: async () => {
      if (!id) return null
      const response = await api.get<{ data: Therapist }>(`/therapists/${id}`)
      return response.data?.data
    },
    enabled: !!id,
  })
}

export function useTherapistStats() {
  return useQuery({
    queryKey: therapistKeys.stats(),
    queryFn: async () => {
      const response = await api.get<TherapistStatsResponse>('/therapists/stats')
      return response.data?.data
    },
  })
}

interface CreateTherapistInput {
  name: string
  email?: string
  phone?: string
  specialty?: string
  user_id?: string
  working_hours?: Record<string, WorkingHours>
  session_duration_minutes?: number
  default_price_cents?: number
  timezone?: string
}

export function useCreateTherapist() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (data: CreateTherapistInput) => {
      const response = await api.post<{ data: Therapist }>('/therapists', data)
      return response.data?.data
    },
    onSuccess: (therapist) => {
      queryClient.invalidateQueries({ queryKey: therapistKeys.lists() })
      queryClient.invalidateQueries({ queryKey: therapistKeys.stats() })
      success('Terapeuta criado', `${therapist.name} foi adicionado com sucesso.`)
    },
    onError: (err) => {
      error('Erro ao criar terapeuta', getErrorMessage(err))
    },
  })
}

interface UpdateTherapistInput extends Partial<CreateTherapistInput> {
  is_active?: boolean
}

export function useUpdateTherapist() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdateTherapistInput }) => {
      const response = await api.put<{ data: Therapist }>(`/therapists/${id}`, data)
      return response.data?.data
    },
    // Optimistic update
    onMutate: async ({ id, data }) => {
      await queryClient.cancelQueries({ queryKey: therapistKeys.detail(id) })
      const previousTherapist = queryClient.getQueryData<Therapist>(therapistKeys.detail(id))

      if (previousTherapist) {
        queryClient.setQueryData<Therapist>(therapistKeys.detail(id), {
          ...previousTherapist,
          ...data,
        })
      }

      return { previousTherapist }
    },
    onError: (err, { id }, context) => {
      if (context?.previousTherapist) {
        queryClient.setQueryData(therapistKeys.detail(id), context.previousTherapist)
      }
      error('Erro ao atualizar terapeuta', getErrorMessage(err))
    },
    onSuccess: (therapist) => {
      success('Terapeuta atualizado', `${therapist.name} foi atualizado com sucesso.`)
    },
    onSettled: (_, __, { id }) => {
      queryClient.invalidateQueries({ queryKey: therapistKeys.lists() })
      queryClient.invalidateQueries({ queryKey: therapistKeys.detail(id) })
      queryClient.invalidateQueries({ queryKey: therapistKeys.stats() })
    },
  })
}

export function useDeleteTherapist() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/therapists/${id}`)
      return id
    },
    // Optimistic update
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: therapistKeys.lists() })

      const previousLists = queryClient.getQueriesData<TherapistsData>({
        queryKey: therapistKeys.lists(),
      })

      queryClient.setQueriesData<TherapistsData>(
        { queryKey: therapistKeys.lists() },
        (old) => {
          if (!old) return old
          return {
            ...old,
            therapists: old.therapists.filter((t) => t.id !== id),
            total: old.total - 1,
          }
        }
      )

      return { previousLists }
    },
    onError: (err, _, context) => {
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data)
        })
      }
      error('Erro ao eliminar terapeuta', getErrorMessage(err))
    },
    onSuccess: () => {
      success('Terapeuta eliminado', 'O terapeuta foi removido com sucesso.')
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: therapistKeys.lists() })
      queryClient.invalidateQueries({ queryKey: therapistKeys.stats() })
    },
  })
}
