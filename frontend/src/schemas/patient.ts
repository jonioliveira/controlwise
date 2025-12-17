import { z } from 'zod'

// Patient schema - patient is linked to client, name/email/phone come from client
export const patientSchema = z.object({
  client_id: z
    .string()
    .min(1, 'Cliente é obrigatório')
    .uuid('ID de cliente inválido'),
  date_of_birth: z
    .string()
    .optional()
    .refine(
      (val) => {
        if (!val) return true
        const date = new Date(val)
        return !isNaN(date.getTime()) && date < new Date()
      },
      { message: 'Data de nascimento inválida' }
    ),
  notes: z
    .string()
    .max(5000, 'Notas devem ter no máximo 5000 caracteres')
    .optional(),
  emergency_contact: z
    .string()
    .max(200, 'Contacto de emergência deve ter no máximo 200 caracteres')
    .optional(),
  emergency_phone: z
    .string()
    .max(20, 'Telefone de emergência deve ter no máximo 20 caracteres')
    .regex(/^[+]?[0-9\s-]*$/, 'Telefone de emergência inválido')
    .optional()
    .or(z.literal('')),
})

// Schema for updating patient (healthcare fields only)
export const patientUpdateSchema = z.object({
  date_of_birth: z
    .string()
    .optional()
    .refine(
      (val) => {
        if (!val) return true
        const date = new Date(val)
        return !isNaN(date.getTime()) && date < new Date()
      },
      { message: 'Data de nascimento inválida' }
    ),
  notes: z
    .string()
    .max(5000, 'Notas devem ter no máximo 5000 caracteres')
    .optional(),
  emergency_contact: z
    .string()
    .max(200, 'Contacto de emergência deve ter no máximo 200 caracteres')
    .optional(),
  emergency_phone: z
    .string()
    .max(20, 'Telefone de emergência deve ter no máximo 20 caracteres')
    .regex(/^[+]?[0-9\s-]*$/, 'Telefone de emergência inválido')
    .optional()
    .or(z.literal('')),
  is_active: z.boolean().optional(),
})

export type PatientFormData = z.infer<typeof patientSchema>
export type PatientUpdateFormData = z.infer<typeof patientUpdateSchema>
