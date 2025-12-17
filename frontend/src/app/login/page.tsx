'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import Link from 'next/link'
import { Building2 } from 'lucide-react'
import { loginSchema, type LoginFormData } from '@/schemas/auth'
import { useLogin } from '@/hooks/useAuth'
import { getErrorMessage } from '@/lib/errors'
import { FormField, FormError, SubmitButton } from '@/components/forms/FormField'

export default function LoginPage() {
  const loginMutation = useLogin()

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data: LoginFormData) => {
    try {
      await loginMutation.mutateAsync(data)
    } catch {
      // Error is handled by the mutation
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-primary-100 flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <Link href="/" className="inline-flex items-center space-x-2 mb-6">
            <Building2 className="h-10 w-10 text-primary-600" />
            <span className="text-3xl font-bold text-gray-900">ControlWise</span>
          </Link>
          <h1 className="text-2xl font-bold text-gray-900">Entrar na sua conta</h1>
          <p className="text-gray-600 mt-2">
            Ou{' '}
            <Link href="/register" className="text-primary-600 hover:text-primary-700 font-medium">
              criar uma nova conta
            </Link>
          </p>
        </div>

        <div className="card">
          <FormError message={loginMutation.error ? getErrorMessage(loginMutation.error) : undefined} />

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              id="email"
              type="email"
              label="Email"
              error={errors.email?.message}
              {...register('email')}
            />

            <FormField
              id="password"
              type="password"
              label="Palavra-passe"
              error={errors.password?.message}
              {...register('password')}
            />

            <div className="flex items-center justify-between text-sm">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                />
                <span className="ml-2 text-gray-700">Lembrar-me</span>
              </label>
              <Link href="/forgot-password" className="text-primary-600 hover:text-primary-700">
                Esqueceu-se da palavra-passe?
              </Link>
            </div>

            <SubmitButton isLoading={isSubmitting || loginMutation.isPending} loadingText="A entrar...">
              Entrar
            </SubmitButton>
          </form>
        </div>
      </div>
    </div>
  )
}
