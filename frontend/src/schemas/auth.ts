import { z } from 'zod'

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'Email é obrigatório')
    .email('Email inválido'),
  password: z
    .string()
    .min(1, 'Palavra-passe é obrigatória'),
})

export type LoginFormData = z.infer<typeof loginSchema>

export const registerSchema = z.object({
  organization_name: z
    .string()
    .min(2, 'Nome da empresa deve ter pelo menos 2 caracteres')
    .max(100, 'Nome da empresa deve ter no máximo 100 caracteres'),
  first_name: z
    .string()
    .min(2, 'Primeiro nome deve ter pelo menos 2 caracteres')
    .max(50, 'Primeiro nome deve ter no máximo 50 caracteres'),
  last_name: z
    .string()
    .min(2, 'Último nome deve ter pelo menos 2 caracteres')
    .max(50, 'Último nome deve ter no máximo 50 caracteres'),
  email: z
    .string()
    .min(1, 'Email é obrigatório')
    .email('Email inválido'),
  phone: z
    .string()
    .min(9, 'Telefone deve ter pelo menos 9 dígitos')
    .max(20, 'Telefone deve ter no máximo 20 caracteres'),
  password: z
    .string()
    .min(8, 'Palavra-passe deve ter pelo menos 8 caracteres')
    .regex(/[A-Z]/, 'Palavra-passe deve conter pelo menos uma maiúscula')
    .regex(/[a-z]/, 'Palavra-passe deve conter pelo menos uma minúscula')
    .regex(/[0-9]/, 'Palavra-passe deve conter pelo menos um número'),
  confirmPassword: z
    .string()
    .min(1, 'Confirmação de palavra-passe é obrigatória'),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'As palavras-passe não coincidem',
  path: ['confirmPassword'],
})

export type RegisterFormData = z.infer<typeof registerSchema>
