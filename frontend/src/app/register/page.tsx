'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Link from 'next/link'
import { Building2 } from 'lucide-react'
import { registerSchema, type RegisterFormData } from '@/schemas/auth'
import { useRegister } from '@/hooks/useAuth'
import { getErrorMessage } from '@/lib/errors'
import { FormField, FormError, SubmitButton } from '@/components/forms/FormField'

export default function RegisterPage() {
  const registerMutation = useRegister()

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = async (data: RegisterFormData) => {
    try {
      const { confirmPassword, ...registerData } = data
      await registerMutation.mutateAsync(registerData)
    } catch {
      // Error is handled by the mutation
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-primary-100 flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-2xl">
        <div className="text-center mb-8">
          <Link href="/" className="inline-flex items-center space-x-2 mb-6">
            <Building2 className="h-10 w-10 text-primary-600" />
            <span className="text-3xl font-bold text-gray-900">ControleWise</span>
          </Link>
          <h1 className="text-2xl font-bold text-gray-900">Criar nova conta</h1>
          <p className="text-gray-600 mt-2">
            Ou{' '}
            <Link href="/login" className="text-primary-600 hover:text-primary-700 font-medium">
              entrar na sua conta existente
            </Link>
          </p>
        </div>

        <div className="card">
          <FormError
            message={registerMutation.error ? getErrorMessage(registerMutation.error) : undefined}
          />

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              id="organization_name"
              type="text"
              label="Nome da Empresa"
              required
              error={errors.organization_name?.message}
              {...register('organization_name')}
            />

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                id="first_name"
                type="text"
                label="Primeiro Nome"
                required
                error={errors.first_name?.message}
                {...register('first_name')}
              />

              <FormField
                id="last_name"
                type="text"
                label="Último Nome"
                required
                error={errors.last_name?.message}
                {...register('last_name')}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                id="email"
                type="email"
                label="Email"
                required
                error={errors.email?.message}
                {...register('email')}
              />

              <FormField
                id="phone"
                type="tel"
                label="Telefone"
                required
                error={errors.phone?.message}
                {...register('phone')}
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                id="password"
                type="password"
                label="Palavra-passe"
                required
                helperText="Mín. 8 caracteres, com maiúscula, minúscula e número"
                error={errors.password?.message}
                {...register('password')}
              />

              <FormField
                id="confirmPassword"
                type="password"
                label="Confirmar Palavra-passe"
                required
                error={errors.confirmPassword?.message}
                {...register('confirmPassword')}
              />
            </div>

            <div className="flex items-start">
              <input
                id="terms"
                type="checkbox"
                required
                className="mt-1 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
              <label htmlFor="terms" className="ml-2 text-sm text-gray-700">
                Concordo com os{' '}
                <Link href="/terms" className="text-primary-600 hover:text-primary-700">
                  Termos de Serviço
                </Link>{' '}
                e{' '}
                <Link href="/privacy" className="text-primary-600 hover:text-primary-700">
                  Política de Privacidade
                </Link>
              </label>
            </div>

            <SubmitButton
              isLoading={isSubmitting || registerMutation.isPending}
              loadingText="A criar conta..."
            >
              Criar Conta
            </SubmitButton>
          </form>
        </div>
      </div>
    </div>
  )
}
