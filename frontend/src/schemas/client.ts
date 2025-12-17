import { z } from 'zod'

export const createClientSchema = z.object({
  name: z
    .string()
    .min(2, 'Nome deve ter pelo menos 2 caracteres')
    .max(100, 'Nome deve ter no máximo 100 caracteres'),
  email: z
    .string()
    .min(1, 'Email é obrigatório')
    .email('Email inválido'),
  phone: z
    .string()
    .min(9, 'Telefone deve ter pelo menos 9 dígitos')
    .max(20, 'Telefone deve ter no máximo 20 caracteres'),
  address: z
    .string()
    .max(500, 'Morada deve ter no máximo 500 caracteres')
    .optional()
    .or(z.literal('')),
  notes: z
    .string()
    .max(1000, 'Notas devem ter no máximo 1000 caracteres')
    .optional()
    .or(z.literal('')),
})

export type CreateClientFormData = z.infer<typeof createClientSchema>

export const updateClientSchema = createClientSchema.partial()

export type UpdateClientFormData = z.infer<typeof updateClientSchema>
