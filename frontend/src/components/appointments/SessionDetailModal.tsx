'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import { pt } from 'date-fns/locale'
import { Modal } from '@/components/ui/Modal'
import {
  useSession,
  useConfirmSession,
  useCancelSession,
  useCompleteSession,
  useMarkNoShow
} from '@/hooks/useSessions'
import { SessionPaymentCard } from './SessionPaymentCard'
import { getStatusColor, sessionStatusOptions } from '@/schemas/session'
import {
  Loader2,
  Calendar,
  Clock,
  User,
  Stethoscope,
  Phone,
  Mail,
  CheckCircle,
  XCircle,
  AlertCircle,
  Edit,
  Euro
} from 'lucide-react'

interface SessionDetailModalProps {
  isOpen: boolean
  onClose: () => void
  sessionId: string | null
  onEdit?: (sessionId: string) => void
}

export function SessionDetailModal({ isOpen, onClose, sessionId, onEdit }: SessionDetailModalProps) {
  const { data: session, isLoading } = useSession(sessionId)
  const confirmSession = useConfirmSession()
  const cancelSession = useCancelSession()
  const completeSession = useCompleteSession()
  const markNoShow = useMarkNoShow()

  const [showCancelForm, setShowCancelForm] = useState(false)
  const [cancelReason, setCancelReason] = useState('')
  const [isProcessing, setIsProcessing] = useState(false)

  const handleConfirm = async () => {
    if (!sessionId) return
    setIsProcessing(true)
    try {
      await confirmSession.mutateAsync(sessionId)
    } catch (error) {
      console.error('Failed to confirm session:', error)
    } finally {
      setIsProcessing(false)
    }
  }

  const handleCancel = async () => {
    if (!sessionId) return
    setIsProcessing(true)
    try {
      await cancelSession.mutateAsync({ id: sessionId, reason: cancelReason || undefined })
      setShowCancelForm(false)
      setCancelReason('')
    } catch (error) {
      console.error('Failed to cancel session:', error)
    } finally {
      setIsProcessing(false)
    }
  }

  const handleComplete = async () => {
    if (!sessionId) return
    setIsProcessing(true)
    try {
      await completeSession.mutateAsync({ id: sessionId })
    } catch (error) {
      console.error('Failed to complete session:', error)
    } finally {
      setIsProcessing(false)
    }
  }

  const handleNoShow = async () => {
    if (!sessionId) return
    setIsProcessing(true)
    try {
      await markNoShow.mutateAsync(sessionId)
    } catch (error) {
      console.error('Failed to mark as no-show:', error)
    } finally {
      setIsProcessing(false)
    }
  }

  if (isLoading || !session) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} title="Detalhes da Sessão">
        <div className="flex items-center justify-center py-8">
          <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
        </div>
      </Modal>
    )
  }

  const statusLabel = sessionStatusOptions.find(s => s.value === session.status)?.label || session.status

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Detalhes da Sessão" size="md">
      <div className="space-y-6">
        {/* Status Badge */}
        <div className="flex items-center justify-between">
          <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(session.status)}`}>
            {statusLabel}
          </span>
          {onEdit && session.status !== 'cancelled' && session.status !== 'completed' && (
            <button
              onClick={() => onEdit(session.id)}
              className="inline-flex items-center text-sm text-primary-600 hover:text-primary-700"
            >
              <Edit className="h-4 w-4 mr-1" />
              Editar
            </button>
          )}
        </div>

        {/* Session Info */}
        <div className="grid grid-cols-2 gap-4">
          <div className="flex items-start gap-3">
            <Calendar className="h-5 w-5 text-gray-400 mt-0.5" />
            <div>
              <p className="text-sm text-gray-500">Data</p>
              <p className="font-medium">
                {format(new Date(session.scheduled_at), "EEEE, d 'de' MMMM 'de' yyyy", { locale: pt })}
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <Clock className="h-5 w-5 text-gray-400 mt-0.5" />
            <div>
              <p className="text-sm text-gray-500">Hora</p>
              <p className="font-medium">
                {format(new Date(session.scheduled_at), 'HH:mm')} ({session.duration_minutes} min)
              </p>
            </div>
          </div>
        </div>

        {/* Patient Info */}
        <div className="bg-gray-50 rounded-lg p-4">
          <h4 className="text-sm font-medium text-gray-900 mb-3 flex items-center gap-2">
            <User className="h-4 w-4" />
            Paciente
          </h4>
          <div className="space-y-2">
            <p className="font-medium text-gray-900">{session.patient_name}</p>
            {session.patient_phone && (
              <p className="text-sm text-gray-600 flex items-center gap-2">
                <Phone className="h-4 w-4" />
                {session.patient_phone}
              </p>
            )}
            {session.patient_email && (
              <p className="text-sm text-gray-600 flex items-center gap-2">
                <Mail className="h-4 w-4" />
                {session.patient_email}
              </p>
            )}
          </div>
        </div>

        {/* Therapist Info */}
        <div className="bg-gray-50 rounded-lg p-4">
          <h4 className="text-sm font-medium text-gray-900 mb-3 flex items-center gap-2">
            <Stethoscope className="h-4 w-4" />
            Terapeuta
          </h4>
          <p className="font-medium text-gray-900">{session.therapist_name}</p>
        </div>

        {/* Price */}
        <div className="flex items-center gap-3">
          <Euro className="h-5 w-5 text-gray-400" />
          <div>
            <p className="text-sm text-gray-500">Preço</p>
            <p className="font-medium">{(session.price_cents / 100).toFixed(2)} EUR</p>
          </div>
        </div>

        {/* Notes */}
        {session.notes && (
          <div>
            <h4 className="text-sm font-medium text-gray-900 mb-2">Notas</h4>
            <p className="text-sm text-gray-600 bg-gray-50 p-3 rounded-lg">{session.notes}</p>
          </div>
        )}

        {/* Cancel reason */}
        {session.cancel_reason && (
          <div>
            <h4 className="text-sm font-medium text-red-700 mb-2">Motivo do Cancelamento</h4>
            <p className="text-sm text-red-600 bg-red-50 p-3 rounded-lg">{session.cancel_reason}</p>
          </div>
        )}

        {/* Payment Card - Only for completed sessions */}
        {session.status === 'completed' && (
          <SessionPaymentCard
            sessionId={session.id}
            sessionPriceCents={session.price_cents}
            sessionStatus={session.status}
          />
        )}

        {/* Cancel Form */}
        {showCancelForm && (
          <div className="bg-red-50 rounded-lg p-4">
            <h4 className="text-sm font-medium text-red-700 mb-2">Cancelar Sessão</h4>
            <textarea
              value={cancelReason}
              onChange={(e) => setCancelReason(e.target.value)}
              rows={2}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-red-500 focus:ring-red-500 text-sm"
              placeholder="Motivo do cancelamento (opcional)"
            />
            <div className="flex gap-2 mt-3">
              <button
                onClick={handleCancel}
                disabled={isProcessing}
                className="inline-flex items-center px-3 py-1.5 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 disabled:opacity-50"
              >
                {isProcessing && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
                Confirmar Cancelamento
              </button>
              <button
                onClick={() => setShowCancelForm(false)}
                className="px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
              >
                Voltar
              </button>
            </div>
          </div>
        )}

        {/* Action Buttons */}
        {!showCancelForm && session.status !== 'cancelled' && session.status !== 'completed' && (
          <div className="flex flex-wrap gap-2 pt-4 border-t border-gray-200">
            {session.status === 'pending' && (
              <button
                onClick={handleConfirm}
                disabled={isProcessing}
                className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {isProcessing ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <CheckCircle className="h-4 w-4 mr-2" />
                )}
                Confirmar
              </button>
            )}

            {(session.status === 'pending' || session.status === 'confirmed') && (
              <>
                <button
                  onClick={handleComplete}
                  disabled={isProcessing}
                  className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 disabled:opacity-50"
                >
                  {isProcessing ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <CheckCircle className="h-4 w-4 mr-2" />
                  )}
                  Concluir
                </button>

                <button
                  onClick={handleNoShow}
                  disabled={isProcessing}
                  className="inline-flex items-center px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 disabled:opacity-50"
                >
                  {isProcessing ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <AlertCircle className="h-4 w-4 mr-2" />
                  )}
                  Faltou
                </button>

                <button
                  onClick={() => setShowCancelForm(true)}
                  className="inline-flex items-center px-4 py-2 text-sm font-medium text-red-700 bg-red-100 rounded-md hover:bg-red-200"
                >
                  <XCircle className="h-4 w-4 mr-2" />
                  Cancelar
                </button>
              </>
            )}
          </div>
        )}
      </div>
    </Modal>
  )
}
