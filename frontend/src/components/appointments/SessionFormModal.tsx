'use client'

import { useEffect } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { format } from 'date-fns'
import { Modal } from '@/components/ui/Modal'
import { sessionSchema, sessionTypeOptions, type SessionFormData } from '@/schemas/session'
import { useCreateSession, useUpdateSession, useSession } from '@/hooks/useSessions'
import { usePatients } from '@/hooks/usePatients'
import { useTherapists, useTherapist } from '@/hooks/useTherapists'
import { Loader2 } from 'lucide-react'

interface SessionFormModalProps {
  isOpen: boolean
  onClose: () => void
  sessionId?: string | null
  initialDate?: Date | null
}

export function SessionFormModal({ isOpen, onClose, sessionId, initialDate }: SessionFormModalProps) {
  const isEditing = !!sessionId
  const { data: session, isLoading: isLoadingSession } = useSession(sessionId ?? null)
  const { data: patientsData } = usePatients({ is_active: true, limit: 1000 })
  const { data: therapistsData } = useTherapists({ is_active: true })
  const createSession = useCreateSession()
  const updateSession = useUpdateSession()

  const {
    register,
    handleSubmit,
    reset,
    watch,
    control,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<SessionFormData>({
    resolver: zodResolver(sessionSchema),
    defaultValues: {
      therapist_id: '',
      patient_id: '',
      scheduled_at: initialDate ? format(initialDate, "yyyy-MM-dd'T'HH:mm") : '',
      duration_minutes: 60,
      price_cents: 0,
      session_type: 'regular',
      notes: '',
    },
  })

  const selectedTherapistId = watch('therapist_id')
  const { data: selectedTherapist } = useTherapist(selectedTherapistId || null)

  // Update duration and price when therapist changes
  useEffect(() => {
    if (selectedTherapist && !isEditing) {
      setValue('duration_minutes', selectedTherapist.session_duration_minutes)
      setValue('price_cents', selectedTherapist.default_price_cents)
    }
  }, [selectedTherapist, setValue, isEditing])

  useEffect(() => {
    if (session && isEditing) {
      reset({
        therapist_id: session.therapist_id,
        patient_id: session.patient_id,
        scheduled_at: format(new Date(session.scheduled_at), "yyyy-MM-dd'T'HH:mm"),
        duration_minutes: session.duration_minutes,
        price_cents: session.price_cents,
        session_type: session.session_type,
        notes: session.notes || '',
      })
    } else if (!isEditing) {
      reset({
        therapist_id: '',
        patient_id: '',
        scheduled_at: initialDate ? format(initialDate, "yyyy-MM-dd'T'HH:mm") : '',
        duration_minutes: 60,
        price_cents: 0,
        session_type: 'regular',
        notes: '',
      })
    }
  }, [session, isEditing, reset, initialDate])

  const onSubmit = async (data: SessionFormData) => {
    try {
      // Convert local datetime to ISO string
      const scheduled_at = new Date(data.scheduled_at).toISOString()

      const payload = {
        ...data,
        scheduled_at,
        notes: data.notes || undefined,
      }

      if (isEditing && sessionId) {
        await updateSession.mutateAsync({ id: sessionId, data: payload })
      } else {
        await createSession.mutateAsync(payload)
      }
      onClose()
    } catch (error) {
      console.error('Failed to save session:', error)
    }
  }

  if (isEditing && isLoadingSession) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} title="Carregar Sessão">
        <div className="flex items-center justify-center py-8">
          <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
        </div>
      </Modal>
    )
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? 'Editar Sessão' : 'Nova Sessão'}
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Terapeuta *
            </label>
            <select
              {...register('therapist_id')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              <option value="">Selecionar terapeuta</option>
              {therapistsData?.therapists?.map((therapist) => (
                <option key={therapist.id} value={therapist.id}>
                  {therapist.name}
                </option>
              ))}
            </select>
            {errors.therapist_id && (
              <p className="mt-1 text-sm text-red-600">{errors.therapist_id.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Paciente *
            </label>
            <select
              {...register('patient_id')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              <option value="">Selecionar paciente</option>
              {patientsData?.patients?.map((patient) => (
                <option key={patient.id} value={patient.id}>
                  {patient.client_name}
                </option>
              ))}
            </select>
            {errors.patient_id && (
              <p className="mt-1 text-sm text-red-600">{errors.patient_id.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Data e Hora *
            </label>
            <input
              type="datetime-local"
              {...register('scheduled_at')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            />
            {errors.scheduled_at && (
              <p className="mt-1 text-sm text-red-600">{errors.scheduled_at.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Tipo de Sessão
            </label>
            <select
              {...register('session_type')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              {sessionTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            {errors.session_type && (
              <p className="mt-1 text-sm text-red-600">{errors.session_type.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Duração (minutos)
            </label>
            <Controller
              name="duration_minutes"
              control={control}
              render={({ field }) => (
                <input
                  type="number"
                  {...field}
                  onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  min={15}
                  max={480}
                />
              )}
            />
            {errors.duration_minutes && (
              <p className="mt-1 text-sm text-red-600">{errors.duration_minutes.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Preço (cêntimos)
            </label>
            <Controller
              name="price_cents"
              control={control}
              render={({ field }) => (
                <input
                  type="number"
                  {...field}
                  onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  min={0}
                />
              )}
            />
            {errors.price_cents && (
              <p className="mt-1 text-sm text-red-600">{errors.price_cents.message}</p>
            )}
            <p className="mt-1 text-xs text-gray-500">
              {((watch('price_cents') || 0) / 100).toFixed(2)} EUR
            </p>
          </div>

          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Notas
            </label>
            <textarea
              {...register('notes')}
              rows={3}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="Notas da sessão..."
            />
            {errors.notes && (
              <p className="mt-1 text-sm text-red-600">{errors.notes.message}</p>
            )}
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50"
          >
            {isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
            {isEditing ? 'Guardar' : 'Criar Sessão'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
