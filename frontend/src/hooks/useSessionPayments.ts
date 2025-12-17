'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/Toast'
import { getErrorMessage } from '@/lib/api-error'
import type {
  SessionPayment,
  SessionPaymentWithDetails,
  SessionPaymentStats,
  UpdateSessionPaymentRequest,
  MarkAsPaidRequest,
} from '@/types'

// Query keys factory
export const sessionPaymentKeys = {
  all: ['sessionPayments'] as const,
  lists: () => [...sessionPaymentKeys.all, 'list'] as const,
  list: (filters: UnpaidFilters) => [...sessionPaymentKeys.lists(), filters] as const,
  details: () => [...sessionPaymentKeys.all, 'detail'] as const,
  detail: (sessionId: string) => [...sessionPaymentKeys.details(), sessionId] as const,
  stats: (filters: StatsFilters) => [...sessionPaymentKeys.all, 'stats', filters] as const,
  patientPayments: (patientId: string) => [...sessionPaymentKeys.all, 'patient', patientId] as const,
}

interface UnpaidFilters {
  therapist_id?: string
  patient_id?: string
  start_date?: string
  end_date?: string
  limit?: number
  offset?: number
}

interface StatsFilters {
  start_date?: string
  end_date?: string
}

interface UnpaidPaymentsResponse {
  data: {
    payments: SessionPaymentWithDetails[]
    total: number
    limit: number
    offset: number
  }
}

interface SessionPaymentResponse {
  data: SessionPayment | null
}

interface StatsResponse {
  data: SessionPaymentStats
}

interface PatientPaymentsResponse {
  data: {
    payments: SessionPaymentWithDetails[]
  }
}

// Get payment for a specific session
export function useSessionPayment(sessionId: string | null) {
  return useQuery({
    queryKey: sessionPaymentKeys.detail(sessionId || ''),
    queryFn: async () => {
      if (!sessionId) return null
      const response = await api.get<SessionPaymentResponse>(`/sessions/${sessionId}/payment`)
      return response.data?.data
    },
    enabled: !!sessionId,
  })
}

// List unpaid session payments
export function useUnpaidPayments(filters: UnpaidFilters = {}) {
  const params = new URLSearchParams()
  if (filters.therapist_id) params.set('therapist_id', filters.therapist_id)
  if (filters.patient_id) params.set('patient_id', filters.patient_id)
  if (filters.start_date) params.set('start_date', filters.start_date)
  if (filters.end_date) params.set('end_date', filters.end_date)
  if (filters.limit) params.set('limit', String(filters.limit))
  if (filters.offset) params.set('offset', String(filters.offset))

  return useQuery({
    queryKey: sessionPaymentKeys.list(filters),
    queryFn: async () => {
      const response = await api.get<UnpaidPaymentsResponse>(`/session-payments/unpaid?${params.toString()}`)
      return response.data?.data
    },
  })
}

// Get payment statistics
export function usePaymentStats(filters: StatsFilters = {}) {
  const params = new URLSearchParams()
  if (filters.start_date) params.set('start_date', filters.start_date)
  if (filters.end_date) params.set('end_date', filters.end_date)

  return useQuery({
    queryKey: sessionPaymentKeys.stats(filters),
    queryFn: async () => {
      const response = await api.get<StatsResponse>(`/session-payments/stats?${params.toString()}`)
      return response.data?.data
    },
  })
}

// Get payment history for a patient
export function usePatientPayments(patientId: string | null) {
  return useQuery({
    queryKey: sessionPaymentKeys.patientPayments(patientId || ''),
    queryFn: async () => {
      if (!patientId) return []
      const response = await api.get<PatientPaymentsResponse>(`/patients/${patientId}/payments`)
      return response.data?.data?.payments || []
    },
    enabled: !!patientId,
  })
}

// Update session payment
export function useUpdateSessionPayment() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ sessionId, data }: { sessionId: string; data: UpdateSessionPaymentRequest }) => {
      const response = await api.put<SessionPaymentResponse>(`/sessions/${sessionId}/payment`, data)
      return response.data?.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: sessionPaymentKeys.all })
      success('Pagamento atualizado', 'As informações de pagamento foram atualizadas.')
    },
    onError: (err) => {
      error('Erro ao atualizar pagamento', getErrorMessage(err))
    },
  })
}

// Mark session payment as paid
export function useMarkSessionAsPaid() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async ({ sessionId, data }: { sessionId: string; data?: MarkAsPaidRequest }) => {
      await api.post(`/sessions/${sessionId}/payment/mark-paid`, data || {})
      return sessionId
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: sessionPaymentKeys.all })
      success('Pagamento registado', 'O pagamento foi marcado como pago.')
    },
    onError: (err) => {
      error('Erro ao registar pagamento', getErrorMessage(err))
    },
  })
}

// Helper function to format cents to currency string
export function formatCentsToEuro(cents: number): string {
  return new Intl.NumberFormat('pt-PT', {
    style: 'currency',
    currency: 'EUR',
  }).format(cents / 100)
}

// Payment status labels
export const paymentStatusLabels: Record<string, string> = {
  unpaid: 'Por pagar',
  partial: 'Parcial',
  paid: 'Pago',
}

// Payment method labels
export const paymentMethodLabels: Record<string, string> = {
  cash: 'Dinheiro',
  transfer: 'Transferência',
  insurance: 'Seguro',
  card: 'Cartão',
}
