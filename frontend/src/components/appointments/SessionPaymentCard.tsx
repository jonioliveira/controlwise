'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import { pt } from 'date-fns/locale'
import {
  useSessionPayment,
  useUpdateSessionPayment,
  useMarkSessionAsPaid,
  formatCentsToEuro,
  paymentStatusLabels,
  paymentMethodLabels,
} from '@/hooks/useSessionPayments'
import type { SessionPaymentStatus, SessionPaymentMethod } from '@/types'
import {
  Loader2,
  CreditCard,
  CheckCircle,
  AlertTriangle,
  Clock,
  Euro,
  Building2,
  Calendar,
  Edit2,
  Save,
  X,
} from 'lucide-react'

interface SessionPaymentCardProps {
  sessionId: string
  sessionPriceCents: number
  sessionStatus: string
  compact?: boolean
}

const paymentStatusConfig: Record<SessionPaymentStatus, { color: string; icon: React.ElementType }> = {
  unpaid: { color: 'text-amber-600 bg-amber-50 border-amber-200', icon: Clock },
  partial: { color: 'text-blue-600 bg-blue-50 border-blue-200', icon: AlertTriangle },
  paid: { color: 'text-green-600 bg-green-50 border-green-200', icon: CheckCircle },
}

const paymentMethods: { value: SessionPaymentMethod; label: string }[] = [
  { value: 'cash', label: 'Dinheiro' },
  { value: 'transfer', label: 'Transferência' },
  { value: 'card', label: 'Cartão' },
  { value: 'insurance', label: 'Seguro' },
]

export function SessionPaymentCard({
  sessionId,
  sessionPriceCents,
  sessionStatus,
  compact = false,
}: SessionPaymentCardProps) {
  const { data: payment, isLoading } = useSessionPayment(sessionId)
  const updatePayment = useUpdateSessionPayment()
  const markAsPaid = useMarkSessionAsPaid()

  const [isEditing, setIsEditing] = useState(false)
  const [editForm, setEditForm] = useState({
    amount_cents: 0,
    payment_method: '' as SessionPaymentMethod | '',
    insurance_provider: '',
    insurance_amount_cents: 0,
    due_date: '',
    notes: '',
  })

  // Determine effective values (use payment data or defaults from session)
  const effectiveAmountCents = payment?.amount_cents ?? sessionPriceCents
  const effectiveStatus: SessionPaymentStatus = payment?.payment_status ?? 'unpaid'
  const statusConfig = paymentStatusConfig[effectiveStatus]
  const StatusIcon = statusConfig.icon

  // Only show payment for completed sessions
  const showPayment = sessionStatus === 'completed'

  if (!showPayment) {
    return null
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-4">
        <Loader2 className="h-5 w-5 animate-spin text-gray-400" />
      </div>
    )
  }

  const handleStartEdit = () => {
    setEditForm({
      amount_cents: effectiveAmountCents,
      payment_method: payment?.payment_method || '',
      insurance_provider: payment?.insurance_provider || '',
      insurance_amount_cents: payment?.insurance_amount_cents || 0,
      due_date: payment?.due_date ? format(new Date(payment.due_date), 'yyyy-MM-dd') : '',
      notes: payment?.notes || '',
    })
    setIsEditing(true)
  }

  const handleSave = async () => {
    await updatePayment.mutateAsync({
      sessionId,
      data: {
        amount_cents: editForm.amount_cents,
        payment_status: effectiveStatus,
        payment_method: editForm.payment_method || undefined,
        insurance_provider: editForm.insurance_provider || undefined,
        insurance_amount_cents: editForm.insurance_amount_cents || undefined,
        due_date: editForm.due_date || undefined,
        notes: editForm.notes || undefined,
      },
    })
    setIsEditing(false)
  }

  const handleMarkAsPaid = async (method?: SessionPaymentMethod) => {
    await markAsPaid.mutateAsync({
      sessionId,
      data: method ? { payment_method: method } : undefined,
    })
  }

  if (compact) {
    return (
      <div className={`flex items-center justify-between p-3 rounded-lg border ${statusConfig.color}`}>
        <div className="flex items-center gap-2">
          <StatusIcon className="h-4 w-4" />
          <span className="text-sm font-medium">{paymentStatusLabels[effectiveStatus]}</span>
          <span className="text-sm">{formatCentsToEuro(effectiveAmountCents)}</span>
        </div>
        {effectiveStatus !== 'paid' && (
          <button
            onClick={() => handleMarkAsPaid()}
            disabled={markAsPaid.isPending}
            className="text-sm font-medium text-green-600 hover:text-green-700 disabled:opacity-50"
          >
            {markAsPaid.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              'Marcar como pago'
            )}
          </button>
        )}
      </div>
    )
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100 bg-gray-50">
        <div className="flex items-center gap-2">
          <CreditCard className="h-5 w-5 text-gray-400" />
          <h4 className="text-sm font-medium text-gray-900">Pagamento</h4>
        </div>
        <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${statusConfig.color}`}>
          <StatusIcon className="h-3.5 w-3.5" />
          {paymentStatusLabels[effectiveStatus]}
        </div>
      </div>

      {/* Content */}
      <div className="p-4 space-y-4">
        {isEditing ? (
          /* Edit Form */
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Valor (EUR)</label>
                <input
                  type="number"
                  step="0.01"
                  value={(editForm.amount_cents / 100).toFixed(2)}
                  onChange={(e) => setEditForm({ ...editForm, amount_cents: Math.round(parseFloat(e.target.value) * 100) })}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Forma de pagamento</label>
                <select
                  value={editForm.payment_method}
                  onChange={(e) => setEditForm({ ...editForm, payment_method: e.target.value as SessionPaymentMethod })}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                >
                  <option value="">Selecionar...</option>
                  {paymentMethods.map((m) => (
                    <option key={m.value} value={m.value}>{m.label}</option>
                  ))}
                </select>
              </div>
            </div>

            {editForm.payment_method === 'insurance' && (
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Seguradora</label>
                  <input
                    type="text"
                    value={editForm.insurance_provider}
                    onChange={(e) => setEditForm({ ...editForm, insurance_provider: e.target.value })}
                    className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                    placeholder="Nome da seguradora"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-gray-700 mb-1">Valor coberto (EUR)</label>
                  <input
                    type="number"
                    step="0.01"
                    value={(editForm.insurance_amount_cents / 100).toFixed(2)}
                    onChange={(e) => setEditForm({ ...editForm, insurance_amount_cents: Math.round(parseFloat(e.target.value) * 100) })}
                    className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                  />
                </div>
              </div>
            )}

            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Data de vencimento</label>
              <input
                type="date"
                value={editForm.due_date}
                onChange={(e) => setEditForm({ ...editForm, due_date: e.target.value })}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Notas</label>
              <textarea
                value={editForm.notes}
                onChange={(e) => setEditForm({ ...editForm, notes: e.target.value })}
                rows={2}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                placeholder="Notas adicionais..."
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <button
                onClick={() => setIsEditing(false)}
                className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
              >
                <X className="h-4 w-4 mr-1" />
                Cancelar
              </button>
              <button
                onClick={handleSave}
                disabled={updatePayment.isPending}
                className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 disabled:opacity-50"
              >
                {updatePayment.isPending ? (
                  <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                ) : (
                  <Save className="h-4 w-4 mr-1" />
                )}
                Guardar
              </button>
            </div>
          </div>
        ) : (
          /* Display Mode */
          <>
            {/* Amount */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-gray-600">
                <Euro className="h-4 w-4" />
                <span className="text-sm">Valor</span>
              </div>
              <span className="font-semibold text-gray-900">{formatCentsToEuro(effectiveAmountCents)}</span>
            </div>

            {/* Payment Method */}
            {payment?.payment_method && (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-gray-600">
                  <CreditCard className="h-4 w-4" />
                  <span className="text-sm">Forma de pagamento</span>
                </div>
                <span className="text-sm text-gray-900">{paymentMethodLabels[payment.payment_method]}</span>
              </div>
            )}

            {/* Insurance */}
            {payment?.insurance_provider && (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-gray-600">
                  <Building2 className="h-4 w-4" />
                  <span className="text-sm">Seguradora</span>
                </div>
                <div className="text-right">
                  <span className="text-sm text-gray-900">{payment.insurance_provider}</span>
                  {payment.insurance_amount_cents && (
                    <span className="text-xs text-gray-500 ml-2">
                      ({formatCentsToEuro(payment.insurance_amount_cents)} cobertos)
                    </span>
                  )}
                </div>
              </div>
            )}

            {/* Due Date */}
            {payment?.due_date && (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-gray-600">
                  <Calendar className="h-4 w-4" />
                  <span className="text-sm">Vencimento</span>
                </div>
                <span className="text-sm text-gray-900">
                  {format(new Date(payment.due_date), "d 'de' MMMM 'de' yyyy", { locale: pt })}
                </span>
              </div>
            )}

            {/* Paid At */}
            {payment?.paid_at && (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-gray-600">
                  <CheckCircle className="h-4 w-4" />
                  <span className="text-sm">Pago em</span>
                </div>
                <span className="text-sm text-gray-900">
                  {format(new Date(payment.paid_at), "d 'de' MMMM 'de' yyyy", { locale: pt })}
                </span>
              </div>
            )}

            {/* Notes */}
            {payment?.notes && (
              <div className="pt-2 border-t border-gray-100">
                <p className="text-xs text-gray-500 mb-1">Notas</p>
                <p className="text-sm text-gray-700">{payment.notes}</p>
              </div>
            )}

            {/* Actions */}
            <div className="flex justify-between items-center pt-3 border-t border-gray-100">
              <button
                onClick={handleStartEdit}
                className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900"
              >
                <Edit2 className="h-4 w-4 mr-1" />
                Editar
              </button>

              {effectiveStatus !== 'paid' && (
                <div className="flex gap-2">
                  {paymentMethods.slice(0, 3).map((method) => (
                    <button
                      key={method.value}
                      onClick={() => handleMarkAsPaid(method.value)}
                      disabled={markAsPaid.isPending}
                      className="inline-flex items-center px-2.5 py-1.5 text-xs font-medium text-green-700 bg-green-50 border border-green-200 rounded-md hover:bg-green-100 disabled:opacity-50"
                    >
                      {method.label}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
