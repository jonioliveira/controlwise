import { z } from 'zod'

// Working hours for a single day - can be undefined (day off)
const workingHoursDaySchema = z.object({
  start: z.string().regex(/^([01]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Hora inválida'),
  end: z.string().regex(/^([01]?[0-9]|2[0-3]):[0-5][0-9]$/, 'Hora inválida'),
}).optional()

export const therapistSchema = z.object({
  name: z
    .string()
    .min(2, 'Nome deve ter pelo menos 2 caracteres')
    .max(200, 'Nome deve ter no máximo 200 caracteres'),
  email: z
    .string()
    .email('Email inválido')
    .max(255, 'Email deve ter no máximo 255 caracteres')
    .optional()
    .or(z.literal('')),
  phone: z
    .string()
    .max(20, 'Telefone deve ter no máximo 20 caracteres')
    .regex(/^[+]?[0-9\s-]*$/, 'Telefone inválido')
    .optional()
    .or(z.literal('')),
  specialty: z
    .string()
    .max(100, 'Especialidade deve ter no máximo 100 caracteres')
    .optional(),
  user_id: z.string().uuid('ID de utilizador inválido').optional().or(z.literal('')),
  working_hours: z.object({
    monday: workingHoursDaySchema,
    tuesday: workingHoursDaySchema,
    wednesday: workingHoursDaySchema,
    thursday: workingHoursDaySchema,
    friday: workingHoursDaySchema,
    saturday: workingHoursDaySchema,
    sunday: workingHoursDaySchema,
  }).optional(),
  session_duration_minutes: z
    .number()
    .min(15, 'Duração mínima de 15 minutos')
    .max(480, 'Duração máxima de 8 horas'),
  default_price_cents: z
    .number()
    .min(0, 'Preço não pode ser negativo'),
  timezone: z.string().min(1, 'Fuso horário é obrigatório'),
})

export type TherapistFormData = z.infer<typeof therapistSchema>

// Default working hours for form initialization
export const defaultWorkingHours = {
  monday: { start: '09:00', end: '18:00' },
  tuesday: { start: '09:00', end: '18:00' },
  wednesday: { start: '09:00', end: '18:00' },
  thursday: { start: '09:00', end: '18:00' },
  friday: { start: '09:00', end: '18:00' },
  saturday: undefined,
  sunday: undefined,
}
