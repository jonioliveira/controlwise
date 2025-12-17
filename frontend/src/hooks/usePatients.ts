'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/Toast'
import { getErrorMessage } from '@/lib/api-error'
import type { Patient } from '@/types'

// Query keys factory for consistent key management
export const patientKeys = {
  all: ['patients'] as const,
  lists: () => [...patientKeys.all, 'list'] as const,
  list: (filters: PatientFilters) => [...patientKeys.lists(), filters] as const,
  details: () => [...patientKeys.all, 'detail'] as const,
  detail: (id: string) => [...patientKeys.details(), id] as const,
  stats: () => [...patientKeys.all, 'stats'] as const,
}

interface PatientFilters {
  search?: string
  is_active?: boolean
  page?: number
  limit?: number
}

interface PatientsResponse {
  data: {
    patients: Patient[]
    total: number
    page: number
    limit: number
  }
}

// Type for cached data (after unwrapping from API response)
interface PatientsData {
  patients: Patient[]
  total: number
  page: number
  limit: number
}

interface PatientResponse {
  data: {
    patient: Patient
  }
}

interface PatientStatsResponse {
  data: {
    total: number
    active: number
    inactive: number
    new_this_month: number
  }
}

export function usePatients(filters: PatientFilters = {}) {
  const params = new URLSearchParams()
  if (filters.search) params.set('search', filters.search)
  if (filters.is_active !== undefined) params.set('is_active', String(filters.is_active))
  if (filters.page) params.set('page', String(filters.page))
  if (filters.limit) params.set('limit', String(filters.limit))

  return useQuery({
    queryKey: patientKeys.list(filters),
    queryFn: async () => {
      const response = await api.get<PatientsResponse>(`/patients?${params.toString()}`)
      return response.data?.data
    },
  })
}

export function usePatient(id: string | null) {
  return useQuery({
    queryKey: patientKeys.detail(id || ''),
    queryFn: async () => {
      if (!id) return null
      const response = await api.get<PatientResponse>(`/patients/${id}`)
      return response.data?.data?.patient
    },
    enabled: !!id,
  })
}

export function usePatientStats() {
  return useQuery({
    queryKey: patientKeys.stats(),
    queryFn: async () => {
      const response = await api.get<PatientStatsResponse>('/patients/stats')
      return response.data?.data
    },
  })
}

// Patient is linked to client - name/email/phone come from client
interface CreatePatientInput {
  client_id: string // Required - link to client
  date_of_birth?: string
  notes?: string // Medical notes
  emergency_contact?: string
  emergency_phone?: string
}

export function useCreatePatient() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (data: CreatePatientInput) => {
      const response = await api.post<PatientResponse>('/patients', data)
      return response.data?.data?.patient
    },
    onSuccess: (patient) => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: patientKeys.lists() })
      queryClient.invalidateQueries({ queryKey: patientKeys.stats() })
      success('Paciente criado', `${patient.client_name} foi adicionado com sucesso.`)
    },
    onError: (err) => {
      error('Erro ao criar paciente', getErrorMessage(err))
    },
  })
}

// Only healthcare fields can be updated - client link cannot be changed
interface UpdatePatientInput {
  date_of_birth?: string
  notes?: string
  emergency_contact?: string
  emergency_phone?: string
  is_active?: boolean
}

export function useUpdatePatient() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: UpdatePatientInput }) => {
      const response = await api.put<PatientResponse>(`/patients/${id}`, data)
      return response.data?.data?.patient
    },
    // Optimistic update
    onMutate: async ({ id, data }) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: patientKeys.detail(id) })
      await queryClient.cancelQueries({ queryKey: patientKeys.lists() })

      // Snapshot the previous value
      const previousPatient = queryClient.getQueryData<Patient>(patientKeys.detail(id))

      // Optimistically update to the new value
      if (previousPatient) {
        queryClient.setQueryData<Patient>(patientKeys.detail(id), {
          ...previousPatient,
          ...data,
        })
      }

      // Return a context object with the snapshotted value
      return { previousPatient }
    },
    onError: (err, { id }, context) => {
      // Rollback on error
      if (context?.previousPatient) {
        queryClient.setQueryData(patientKeys.detail(id), context.previousPatient)
      }
      error('Erro ao atualizar paciente', getErrorMessage(err))
    },
    onSuccess: (patient) => {
      success('Paciente atualizado', `${patient.client_name} foi atualizado com sucesso.`)
    },
    onSettled: (_, __, { id }) => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: patientKeys.lists() })
      queryClient.invalidateQueries({ queryKey: patientKeys.detail(id) })
      queryClient.invalidateQueries({ queryKey: patientKeys.stats() })
    },
  })
}

export function useDeletePatient() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/patients/${id}`)
      return id
    },
    // Optimistic update
    onMutate: async (id) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: patientKeys.lists() })

      // Snapshot all patient list queries
      const previousLists = queryClient.getQueriesData<PatientsData>({
        queryKey: patientKeys.lists(),
      })

      // Optimistically remove from all lists
      queryClient.setQueriesData<PatientsData>(
        { queryKey: patientKeys.lists() },
        (old) => {
          if (!old) return old
          return {
            ...old,
            patients: old.patients.filter((p) => p.id !== id),
            total: old.total - 1,
          }
        }
      )

      return { previousLists }
    },
    onError: (err, _, context) => {
      // Rollback on error
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data)
        })
      }
      error('Erro ao eliminar paciente', getErrorMessage(err))
    },
    onSuccess: () => {
      success('Paciente eliminado', 'O paciente foi removido com sucesso.')
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: patientKeys.lists() })
      queryClient.invalidateQueries({ queryKey: patientKeys.stats() })
    },
  })
}
