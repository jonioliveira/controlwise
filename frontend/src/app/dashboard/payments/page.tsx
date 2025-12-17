'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import { pt } from 'date-fns/locale'
import {
  CreditCard,
  Search,
  Calendar,
  User,
  Stethoscope,
  Loader2,
  Euro,
  CheckCircle,
  Clock,
  AlertTriangle,
  Filter,
  ChevronLeft,
  ChevronRight,
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
import type { SessionPaymentWithDetails, SessionPaymentMethod } from '@/types'

const paymentMethods: { value: SessionPaymentMethod; label: string }[] = [
  { value: 'cash', label: 'Dinheiro' },
  { value: 'transfer', label: 'Transferência' },
  { value: 'card', label: 'Cartão' },
  { value: 'insurance', label: 'Seguro' },
]

export default function PaymentsPage() {
  const [filters, setFilters] = useState({
    therapist_id: '',
    start_date: '',
    end_date: '',
    offset: 0,
    limit: 20,
  })
  const [showFilters, setShowFilters] = useState(false)
  const [processingId, setProcessingId] = useState<string | null>(null)

  const { data, isLoading, error } = useUnpaidPayments(filters)
  const { data: stats, isLoading: statsLoading } = usePaymentStats({
    start_date: filters.start_date,
    end_date: filters.end_date,
  })
  const { data: therapistsData } = useTherapists({ limit: 100 })
  const markAsPaid = useMarkSessionAsPaid()

  const therapists = therapistsData?.therapists || []
  const payments = data?.payments || []
  const total = data?.total || 0
  const currentPage = Math.floor(filters.offset / filters.limit) + 1
  const totalPages = Math.ceil(total / filters.limit)

  const handleMarkAsPaid = async (sessionId: string, method: SessionPaymentMethod) => {
    setProcessingId(sessionId)
    try {
      await markAsPaid.mutateAsync({ sessionId, data: { payment_method: method } })
    } catch (error) {
      console.error('Failed to mark as paid:', error)
    } finally {
      setProcessingId(null)
    }
  }

  const handlePrevPage = () => {
    if (filters.offset > 0) {
      setFilters({ ...filters, offset: filters.offset - filters.limit })
    }
  }

  const handleNextPage = () => {
    if (filters.offset + filters.limit < total) {
      setFilters({ ...filters, offset: filters.offset + filters.limit })
    }
  }

  const clearFilters = () => {
    setFilters({
      therapist_id: '',
      start_date: '',
      end_date: '',
      offset: 0,
      limit: 20,
    })
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Pagamentos</h1>
          <p className="mt-1 text-sm text-gray-500">
            Gerir pagamentos de sessões e visualizar pendentes
          </p>
        </div>
      </div>

      {/* Stats Cards */}
      {!statsLoading && stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-100">
                <Clock className="h-5 w-5 text-amber-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Por Pagar</p>
                <p className="text-xl font-semibold text-gray-900">{stats.unpaid_count}</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
                <AlertTriangle className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Parciais</p>
                <p className="text-xl font-semibold text-gray-900">{stats.partial_count}</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100">
                <CheckCircle className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Pagos</p>
                <p className="text-xl font-semibold text-gray-900">{stats.paid_count}</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-red-100">
                <Euro className="h-5 w-5 text-red-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Total Pendente</p>
                <p className="text-xl font-semibold text-gray-900">{formatCentsToEuro(stats.total_unpaid_cents)}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div className="flex items-center justify-between mb-4">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900"
          >
            <Filter className="h-4 w-4 mr-2" />
            Filtros
            {(filters.therapist_id || filters.start_date || filters.end_date) && (
              <span className="ml-2 px-2 py-0.5 text-xs bg-primary-100 text-primary-700 rounded-full">
                Ativos
              </span>
            )}
          </button>
          {(filters.therapist_id || filters.start_date || filters.end_date) && (
            <button
              onClick={clearFilters}
              className="text-sm text-primary-600 hover:text-primary-700"
            >
              Limpar filtros
            </button>
          )}
        </div>

        {showFilters && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-4 border-t border-gray-100">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Terapeuta</label>
              <select
                value={filters.therapist_id}
                onChange={(e) => setFilters({ ...filters, therapist_id: e.target.value, offset: 0 })}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
              >
                <option value="">Todos os terapeutas</option>
                {therapists.map((t) => (
                  <option key={t.id} value={t.id}>{t.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Data inicial</label>
              <input
                type="date"
                value={filters.start_date}
                onChange={(e) => setFilters({ ...filters, start_date: e.target.value, offset: 0 })}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Data final</label>
              <input
                type="date"
                value={filters.end_date}
                onChange={(e) => setFilters({ ...filters, end_date: e.target.value, offset: 0 })}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
              />
            </div>
          </div>
        )}
      </div>

      {/* Payments List */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
          </div>
        ) : error ? (
          <div className="text-center py-12 text-red-600">
            Erro ao carregar pagamentos. Tente novamente.
          </div>
        ) : payments.length === 0 ? (
          <div className="text-center py-12">
            <CreditCard className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">Sem pagamentos pendentes</h3>
            <p className="mt-1 text-sm text-gray-500">
              Todos os pagamentos estao em dia.
            </p>
          </div>
        ) : (
          <>
            {/* Table */}
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Sessao
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Paciente
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Terapeuta
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Valor
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Estado
                    </th>
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Acoes
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {payments.map((payment) => (
                    <PaymentRow
                      key={payment.session_id}
                      payment={payment}
                      onMarkAsPaid={handleMarkAsPaid}
                      isProcessing={processingId === payment.session_id}
                    />
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between px-6 py-4 border-t border-gray-200">
                <div className="text-sm text-gray-500">
                  Mostrando {filters.offset + 1} a {Math.min(filters.offset + filters.limit, total)} de {total} pagamentos
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={handlePrevPage}
                    disabled={currentPage === 1}
                    className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </button>
                  <span className="text-sm text-gray-700">
                    Pagina {currentPage} de {totalPages}
                  </span>
                  <button
                    onClick={handleNextPage}
                    disabled={currentPage === totalPages}
                    className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <ChevronRight className="h-4 w-4" />
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

interface PaymentRowProps {
  payment: SessionPaymentWithDetails
  onMarkAsPaid: (sessionId: string, method: SessionPaymentMethod) => void
  isProcessing: boolean
}

function PaymentRow({ payment, onMarkAsPaid, isProcessing }: PaymentRowProps) {
  const [showPaymentOptions, setShowPaymentOptions] = useState(false)

  return (
    <tr className="hover:bg-gray-50">
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center gap-2">
          <Calendar className="h-4 w-4 text-gray-400" />
          <div>
            <p className="text-sm font-medium text-gray-900">
              {format(new Date(payment.scheduled_at), "d MMM yyyy", { locale: pt })}
            </p>
            <p className="text-xs text-gray-500">
              {format(new Date(payment.scheduled_at), "HH:mm")}
            </p>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center gap-2">
          <User className="h-4 w-4 text-gray-400" />
          <span className="text-sm text-gray-900">{payment.patient_name}</span>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center gap-2">
          <Stethoscope className="h-4 w-4 text-gray-400" />
          <span className="text-sm text-gray-900">{payment.therapist_name}</span>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <span className="text-sm font-semibold text-gray-900">
          {formatCentsToEuro(payment.amount_cents)}
        </span>
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
          payment.payment_status === 'paid'
            ? 'bg-green-100 text-green-800'
            : payment.payment_status === 'partial'
            ? 'bg-blue-100 text-blue-800'
            : 'bg-amber-100 text-amber-800'
        }`}>
          {paymentStatusLabels[payment.payment_status]}
        </span>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-right">
        {payment.payment_status !== 'paid' && (
          <div className="relative">
            <button
              onClick={() => setShowPaymentOptions(!showPaymentOptions)}
              disabled={isProcessing}
              className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-green-700 bg-green-50 border border-green-200 rounded-md hover:bg-green-100 disabled:opacity-50"
            >
              {isProcessing ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <>
                  <CheckCircle className="h-4 w-4 mr-1" />
                  Pagar
                </>
              )}
            </button>

            {showPaymentOptions && !isProcessing && (
              <>
                <div
                  className="fixed inset-0 z-10"
                  onClick={() => setShowPaymentOptions(false)}
                />
                <div className="absolute right-0 top-full mt-1 w-40 bg-white rounded-md shadow-lg border border-gray-200 z-20">
                  <div className="py-1">
                    {paymentMethods.map((method) => (
                      <button
                        key={method.value}
                        onClick={() => {
                          onMarkAsPaid(payment.session_id, method.value)
                          setShowPaymentOptions(false)
                        }}
                        className="block w-full px-4 py-2 text-sm text-left text-gray-700 hover:bg-gray-100"
                      >
                        {method.label}
                      </button>
                    ))}
                  </div>
                </div>
              </>
            )}
          </div>
        )}
      </td>
    </tr>
  )
}
