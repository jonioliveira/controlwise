'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import { pt } from 'date-fns/locale'
import {
  CreditCard,
  Loader2,
  CheckCircle,
  Clock,
  AlertTriangle,
  Euro,
  User,
  Stethoscope,
  Calendar,
  Filter,
  TrendingUp,
  TrendingDown,
} from 'lucide-react'
import {
  useUnpaidPayments,
  usePaymentStats,
  useMarkSessionAsPaid,
  formatCentsToEuro,
  paymentStatusLabels,
  paymentMethodLabels,
} from '@/hooks/useSessionPayments'
import { useTherapists } from '@/hooks/useTherapists'
import type { SessionPaymentStatus, SessionPaymentMethod, SessionPaymentWithDetails } from '@/types'

const paymentStatusConfig: Record<SessionPaymentStatus, { color: string; bgColor: string; icon: React.ElementType }> = {
  unpaid: { color: 'text-amber-600', bgColor: 'bg-amber-50 border-amber-200', icon: Clock },
  partial: { color: 'text-blue-600', bgColor: 'bg-blue-50 border-blue-200', icon: AlertTriangle },
  paid: { color: 'text-green-600', bgColor: 'bg-green-50 border-green-200', icon: CheckCircle },
}

const paymentMethods: { value: SessionPaymentMethod; label: string }[] = [
  { value: 'cash', label: 'Dinheiro' },
  { value: 'transfer', label: 'Transferência' },
  { value: 'card', label: 'Cartão' },
  { value: 'insurance', label: 'Seguro' },
]

export default function SessionPaymentsPage() {
  const [filters, setFilters] = useState({
    therapist_id: '',
    start_date: '',
    end_date: '',
  })
  const [showFilters, setShowFilters] = useState(false)

  const { data: unpaidData, isLoading: isLoadingPayments } = useUnpaidPayments({
    therapist_id: filters.therapist_id || undefined,
    start_date: filters.start_date || undefined,
    end_date: filters.end_date || undefined,
    limit: 50,
  })

  const { data: stats, isLoading: isLoadingStats } = usePaymentStats({
    start_date: filters.start_date || undefined,
    end_date: filters.end_date || undefined,
  })

  const { data: therapistsData } = useTherapists()
  const therapists = therapistsData?.therapists || []
  const markAsPaid = useMarkSessionAsPaid()

  const payments = unpaidData?.payments || []
  const total = unpaidData?.total || 0

  const handleMarkAsPaid = async (sessionId: string, method: SessionPaymentMethod) => {
    await markAsPaid.mutateAsync({ sessionId, data: { payment_method: method } })
  }

  const clearFilters = () => {
    setFilters({
      therapist_id: '',
      start_date: '',
      end_date: '',
    })
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Pagamentos de Sessões</h1>
          <p className="text-sm text-gray-500">
            Gestão de pagamentos de sessões de terapia
          </p>
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md border ${
            showFilters ? 'bg-primary-50 text-primary-700 border-primary-200' : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
          }`}
        >
          <Filter className="h-4 w-4" />
          Filtros
        </button>
      </div>

      {/* Stats Cards */}
      {!isLoadingStats && stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-gray-500">
                <Calendar className="h-4 w-4" />
                <span className="text-sm">Total Sessões</span>
              </div>
            </div>
            <p className="mt-2 text-2xl font-bold text-gray-900">{stats.total_sessions}</p>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-green-600">
                <CheckCircle className="h-4 w-4" />
                <span className="text-sm">Pagas</span>
              </div>
              <TrendingUp className="h-4 w-4 text-green-500" />
            </div>
            <p className="mt-2 text-2xl font-bold text-green-600">{stats.paid_count}</p>
            <p className="text-sm text-gray-500">{formatCentsToEuro(stats.total_paid_cents)}</p>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-amber-600">
                <Clock className="h-4 w-4" />
                <span className="text-sm">Por Pagar</span>
              </div>
              <TrendingDown className="h-4 w-4 text-amber-500" />
            </div>
            <p className="mt-2 text-2xl font-bold text-amber-600">{stats.unpaid_count}</p>
            <p className="text-sm text-gray-500">{formatCentsToEuro(stats.total_unpaid_cents)}</p>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-blue-600">
                <AlertTriangle className="h-4 w-4" />
                <span className="text-sm">Parcial</span>
              </div>
            </div>
            <p className="mt-2 text-2xl font-bold text-blue-600">{stats.partial_count}</p>
          </div>
        </div>
      )}

      {/* Filters */}
      {showFilters && (
        <div className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Terapeuta
              </label>
              <select
                value={filters.therapist_id}
                onChange={(e) => setFilters({ ...filters, therapist_id: e.target.value })}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
              >
                <option value="">Todos</option>
                {therapists.map((t) => (
                  <option key={t.id} value={t.id}>{t.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Data Início
              </label>
              <input
                type="date"
                value={filters.start_date}
                onChange={(e) => setFilters({ ...filters, start_date: e.target.value })}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Data Fim
              </label>
              <input
                type="date"
                value={filters.end_date}
                onChange={(e) => setFilters({ ...filters, end_date: e.target.value })}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
              />
            </div>
            <div className="flex items-end">
              <button
                onClick={clearFilters}
                className="w-full px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
              >
                Limpar Filtros
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Payments List */}
      <div className="bg-white rounded-lg border border-gray-200">
        <div className="px-4 py-3 border-b border-gray-200">
          <h2 className="text-lg font-medium text-gray-900">
            Sessões por Pagar
            {total > 0 && (
              <span className="ml-2 text-sm font-normal text-gray-500">
                ({total} {total === 1 ? 'sessão' : 'sessões'})
              </span>
            )}
          </h2>
        </div>

        {isLoadingPayments ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
          </div>
        ) : payments.length === 0 ? (
          <div className="text-center py-12">
            <CheckCircle className="h-12 w-12 text-green-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900">Tudo em dia!</h3>
            <p className="mt-1 text-sm text-gray-500">
              Não existem sessões pendentes de pagamento
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {payments.map((payment) => (
              <PaymentRow
                key={payment.id}
                payment={payment}
                onMarkAsPaid={handleMarkAsPaid}
                isPending={markAsPaid.isPending}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function PaymentRow({
  payment,
  onMarkAsPaid,
  isPending,
}: {
  payment: SessionPaymentWithDetails
  onMarkAsPaid: (sessionId: string, method: SessionPaymentMethod) => void
  isPending: boolean
}) {
  const statusConfig = paymentStatusConfig[payment.payment_status]
  const StatusIcon = statusConfig.icon

  return (
    <div className="px-4 py-4 hover:bg-gray-50">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {/* Status Badge */}
          <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border ${statusConfig.bgColor} ${statusConfig.color}`}>
            <StatusIcon className="h-3.5 w-3.5" />
            {paymentStatusLabels[payment.payment_status]}
          </div>

          {/* Session Info */}
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-gray-400" />
              <span className="font-medium text-gray-900">{payment.patient_name}</span>
            </div>
            <div className="flex items-center gap-4 mt-1 text-sm text-gray-500">
              <div className="flex items-center gap-1">
                <Stethoscope className="h-3.5 w-3.5" />
                <span>{payment.therapist_name}</span>
              </div>
              <div className="flex items-center gap-1">
                <Calendar className="h-3.5 w-3.5" />
                <span>
                  {format(new Date(payment.scheduled_at), "d 'de' MMM 'de' yyyy 'às' HH:mm", { locale: pt })}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-4">
          {/* Amount */}
          <div className="text-right">
            <div className="flex items-center gap-1 text-lg font-semibold text-gray-900">
              <Euro className="h-4 w-4" />
              {formatCentsToEuro(payment.amount_cents)}
            </div>
            {payment.due_date && (
              <p className="text-xs text-gray-500">
                Vence: {format(new Date(payment.due_date), 'dd/MM/yyyy')}
              </p>
            )}
          </div>

          {/* Actions */}
          {payment.payment_status !== 'paid' && (
            <div className="flex gap-2">
              {paymentMethods.map((method) => (
                <button
                  key={method.value}
                  onClick={() => onMarkAsPaid(payment.session_id, method.value)}
                  disabled={isPending}
                  className="inline-flex items-center px-2.5 py-1.5 text-xs font-medium text-green-700 bg-green-50 border border-green-200 rounded-md hover:bg-green-100 disabled:opacity-50"
                  title={`Marcar como pago com ${method.label}`}
                >
                  {method.label}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Additional Info */}
      {(payment.insurance_provider || payment.notes) && (
        <div className="mt-3 pt-3 border-t border-gray-100 flex items-center gap-4 text-sm text-gray-500">
          {payment.insurance_provider && (
            <span>Seguro: {payment.insurance_provider}</span>
          )}
          {payment.notes && (
            <span className="truncate">Notas: {payment.notes}</span>
          )}
        </div>
      )}
    </div>
  )
}
