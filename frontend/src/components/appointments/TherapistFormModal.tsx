'use client'

import { useEffect } from 'react'
import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Modal } from '@/components/ui/Modal'
import { therapistSchema, defaultWorkingHours, type TherapistFormData } from '@/schemas/therapist'
import { useCreateTherapist, useUpdateTherapist, useTherapist } from '@/hooks/useTherapists'
import { Loader2 } from 'lucide-react'

interface TherapistFormModalProps {
  isOpen: boolean
  onClose: () => void
  therapistId?: string | null
}

type WeekDay = 'monday' | 'tuesday' | 'wednesday' | 'thursday' | 'friday' | 'saturday' | 'sunday'

const weekDays: { key: WeekDay; label: string }[] = [
  { key: 'monday', label: 'Segunda' },
  { key: 'tuesday', label: 'Terça' },
  { key: 'wednesday', label: 'Quarta' },
  { key: 'thursday', label: 'Quinta' },
  { key: 'friday', label: 'Sexta' },
  { key: 'saturday', label: 'Sábado' },
  { key: 'sunday', label: 'Domingo' },
]

export function TherapistFormModal({ isOpen, onClose, therapistId }: TherapistFormModalProps) {
  const isEditing = !!therapistId
  const { data: therapist, isLoading: isLoadingTherapist } = useTherapist(therapistId ?? null)
  const createTherapist = useCreateTherapist()
  const updateTherapist = useUpdateTherapist()

  const {
    register,
    handleSubmit,
    reset,
    control,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<TherapistFormData>({
    resolver: zodResolver(therapistSchema),
    defaultValues: {
      name: '',
      email: '',
      phone: '',
      specialty: '',
      user_id: '',
      working_hours: defaultWorkingHours,
      session_duration_minutes: 60,
      default_price_cents: 5000,
      timezone: 'Europe/Lisbon',
    },
  })

  useEffect(() => {
    if (therapist && isEditing) {
      reset({
        name: therapist.name,
        email: therapist.email || '',
        phone: therapist.phone || '',
        specialty: therapist.specialty || '',
        user_id: therapist.user_id || '',
        working_hours: therapist.working_hours || defaultWorkingHours,
        session_duration_minutes: therapist.session_duration_minutes,
        default_price_cents: therapist.default_price_cents,
        timezone: therapist.timezone,
      })
    } else if (!isEditing) {
      reset({
        name: '',
        email: '',
        phone: '',
        specialty: '',
        user_id: '',
        working_hours: defaultWorkingHours,
        session_duration_minutes: 60,
        default_price_cents: 5000,
        timezone: 'Europe/Lisbon',
      })
    }
  }, [therapist, isEditing, reset])

  const onSubmit = async (data: TherapistFormData) => {
    try {
      const cleanedData = {
        ...data,
        email: data.email || undefined,
        phone: data.phone || undefined,
        specialty: data.specialty || undefined,
        user_id: data.user_id || undefined,
      }

      if (isEditing && therapistId) {
        await updateTherapist.mutateAsync({ id: therapistId, data: cleanedData })
      } else {
        await createTherapist.mutateAsync(cleanedData)
      }
      onClose()
    } catch (error) {
      console.error('Failed to save therapist:', error)
    }
  }

  if (isEditing && isLoadingTherapist) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} title="Carregar Terapeuta">
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
      title={isEditing ? 'Editar Terapeuta' : 'Novo Terapeuta'}
      size="lg"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Basic Info */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Nome *
            </label>
            <input
              type="text"
              {...register('name')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="Nome completo"
            />
            {errors.name && (
              <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Email
            </label>
            <input
              type="email"
              {...register('email')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="email@exemplo.com"
            />
            {errors.email && (
              <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Telefone
            </label>
            <input
              type="tel"
              {...register('phone')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="+351 912 345 678"
            />
            {errors.phone && (
              <p className="mt-1 text-sm text-red-600">{errors.phone.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Especialidade
            </label>
            <input
              type="text"
              {...register('specialty')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="Fisioterapia, Psicologia..."
            />
            {errors.specialty && (
              <p className="mt-1 text-sm text-red-600">{errors.specialty.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Fuso Horário
            </label>
            <select
              {...register('timezone')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              <option value="Europe/Lisbon">Europe/Lisbon</option>
              <option value="Europe/London">Europe/London</option>
              <option value="Europe/Paris">Europe/Paris</option>
              <option value="Europe/Madrid">Europe/Madrid</option>
            </select>
          </div>
        </div>

        {/* Session Settings */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Duração da Sessão (minutos)
            </label>
            <Controller
              name="session_duration_minutes"
              control={control}
              render={({ field }) => (
                <input
                  type="number"
                  {...field}
                  onChange={(e) => field.onChange(parseInt(e.target.value) || 60)}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  min={15}
                  max={480}
                />
              )}
            />
            {errors.session_duration_minutes && (
              <p className="mt-1 text-sm text-red-600">{errors.session_duration_minutes.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Preço por Sessão (cêntimos)
            </label>
            <Controller
              name="default_price_cents"
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
            <p className="mt-1 text-xs text-gray-500">
              {((watch('default_price_cents') || 0) / 100).toFixed(2)} EUR
            </p>
            {errors.default_price_cents && (
              <p className="mt-1 text-sm text-red-600">{errors.default_price_cents.message}</p>
            )}
          </div>
        </div>

        {/* Working Hours */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Horário de Trabalho
          </label>
          <div className="space-y-3">
            {weekDays.map(({ key, label }) => (
              <Controller
                key={key}
                name={`working_hours.${key}`}
                control={control}
                render={({ field }) => {
                  const value = field.value as { start: string; end: string } | undefined
                  return (
                    <div className="flex items-center gap-4">
                      <span className="w-20 text-sm text-gray-600">{label}</span>
                      <input
                        type="time"
                        value={value?.start || ''}
                        onChange={(e) => field.onChange({
                          start: e.target.value,
                          end: value?.end || ''
                        })}
                        className="rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                      />
                      <span className="text-gray-400">-</span>
                      <input
                        type="time"
                        value={value?.end || ''}
                        onChange={(e) => field.onChange({
                          start: value?.start || '',
                          end: e.target.value
                        })}
                        className="rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                      />
                      <button
                        type="button"
                        onClick={() => field.onChange(undefined)}
                        className="text-xs text-gray-500 hover:text-gray-700"
                      >
                        Limpar
                      </button>
                    </div>
                  )
                }}
              />
            ))}
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
            {isEditing ? 'Guardar' : 'Criar Terapeuta'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
