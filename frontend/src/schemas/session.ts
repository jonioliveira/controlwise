import { z } from 'zod'

export const sessionTypeOptions = [
  { value: 'regular', label: 'Regular' },
  { value: 'evaluation', label: 'Avaliação' },
  { value: 'follow_up', label: 'Acompanhamento' },
] as const

export const sessionStatusOptions = [
  { value: 'pending', label: 'Pendente', color: 'yellow' },
  { value: 'confirmed', label: 'Confirmada', color: 'blue' },
  { value: 'cancelled', label: 'Cancelada', color: 'red' },
  { value: 'completed', label: 'Concluída', color: 'green' },
  { value: 'no_show', label: 'Faltou', color: 'gray' },
] as const

export const sessionSchema = z.object({
  therapist_id: z.string().uuid('Selecione um terapeuta'),
  patient_id: z.string().uuid('Selecione um paciente'),
  scheduled_at: z
    .string()
    .min(1, 'Selecione uma data e hora')
    .refine(
      (val) => {
        const date = new Date(val)
        return !isNaN(date.getTime())
      },
      { message: 'Data e hora inválidas' }
    ),
  duration_minutes: z
    .number()
    .min(15, 'Duração mínima de 15 minutos')
    .max(480, 'Duração máxima de 8 horas')
    .optional(),
  price_cents: z
    .number()
    .min(0, 'Preço não pode ser negativo')
    .optional(),
  session_type: z
    .enum(['regular', 'evaluation', 'follow_up'])
    .default('regular'),
  notes: z
    .string()
    .max(5000, 'Notas devem ter no máximo 5000 caracteres')
    .optional(),
})

export type SessionFormData = z.infer<typeof sessionSchema>

export const cancelSessionSchema = z.object({
  reason: z
    .string()
    .max(1000, 'Motivo deve ter no máximo 1000 caracteres')
    .optional(),
})

export type CancelSessionFormData = z.infer<typeof cancelSessionSchema>

// Helper to get status badge styling
export function getStatusColor(status: string): string {
  const statusConfig = sessionStatusOptions.find((s) => s.value === status)
  switch (statusConfig?.color) {
    case 'yellow':
      return 'bg-yellow-100 text-yellow-800'
    case 'blue':
      return 'bg-blue-100 text-blue-800'
    case 'red':
      return 'bg-red-100 text-red-800'
    case 'green':
      return 'bg-green-100 text-green-800'
    case 'gray':
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

// Helper to get calendar event color
export function getCalendarEventColor(status: string): string {
  switch (status) {
    case 'pending':
      return '#f59e0b' // yellow
    case 'confirmed':
      return '#3b82f6' // blue
    case 'cancelled':
      return '#ef4444' // red
    case 'completed':
      return '#22c55e' // green
    case 'no_show':
      return '#6b7280' // gray
    default:
      return '#6b7280'
  }
}
