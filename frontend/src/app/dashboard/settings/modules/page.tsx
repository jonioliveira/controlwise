'use client'

import Link from 'next/link'
import {
  ArrowLeft,
  Puzzle,
  Building2,
  Calendar,
  Bell,
  Check,
  X,
  Loader2,
  ShieldAlert
} from 'lucide-react'
import { useModules } from '@/hooks/useModules'
import type { ModuleName } from '@/types'

const moduleIcons: Record<ModuleName, React.ComponentType<{ className?: string }>> = {
  construction: Building2,
  appointments: Calendar,
  notifications: Bell
}

const moduleDescriptions: Record<ModuleName, string> = {
  construction: 'Gestão de obras, orçamentos, projetos e pagamentos',
  appointments: 'Agenda, sessões, pacientes e terapeutas',
  notifications: 'Notificações por WhatsApp e lembretes automáticos'
}

export default function ModulesSettingsPage() {
  const { data: modules, isLoading, error } = useModules()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-md bg-red-50 p-4">
        <p className="text-sm text-red-700">
          Erro ao carregar módulos. Por favor tente novamente.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href="/dashboard/settings"
          className="flex items-center justify-center h-10 w-10 rounded-lg border border-gray-200 bg-white text-gray-500 hover:bg-gray-50 hover:text-gray-700"
        >
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Módulos</h1>
          <p className="mt-1 text-sm text-gray-500">
            Funcionalidades ativas na sua organização
          </p>
        </div>
      </div>

      {/* Admin-only notice */}
      <div className="rounded-lg bg-amber-50 border border-amber-200 p-4">
        <div className="flex items-start gap-3">
          <ShieldAlert className="h-5 w-5 text-amber-600 mt-0.5" />
          <div>
            <h3 className="text-sm font-medium text-amber-800">
              Gestão de módulos restrita
            </h3>
            <p className="mt-1 text-sm text-amber-700">
              A ativação e desativação de módulos é gerida pelo administrador da plataforma.
              Contacte o suporte se precisar de alterações.
            </p>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        {modules?.map((module) => {
          const Icon = moduleIcons[module.module_name] || Puzzle

          return (
            <div
              key={module.module_name}
              className="flex items-start gap-4 rounded-lg border border-gray-200 bg-white p-5 shadow-sm"
            >
              <div
                className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-lg ${
                  module.is_enabled
                    ? 'bg-primary-50 text-primary-600'
                    : 'bg-gray-100 text-gray-400'
                }`}
              >
                <Icon className="h-6 w-6" />
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <h3 className="text-base font-semibold text-gray-900">
                    {module.display_name}
                  </h3>
                  <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                    module.is_enabled
                      ? 'bg-green-100 text-green-700'
                      : 'bg-gray-100 text-gray-600'
                  }`}>
                    {module.is_enabled ? 'Ativo' : 'Inativo'}
                  </span>
                </div>
                <p className="mt-1 text-sm text-gray-500">
                  {module.description || moduleDescriptions[module.module_name]}
                </p>
              </div>

              <div className="shrink-0">
                <div
                  className={`relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent cursor-not-allowed opacity-50 ${
                    module.is_enabled ? 'bg-primary-600' : 'bg-gray-200'
                  }`}
                >
                  <span
                    className={`pointer-events-none inline-flex h-5 w-5 transform items-center justify-center rounded-full bg-white shadow ring-0 ${
                      module.is_enabled ? 'translate-x-5' : 'translate-x-0'
                    }`}
                  >
                    {module.is_enabled ? (
                      <Check className="h-3 w-3 text-primary-600" />
                    ) : (
                      <X className="h-3 w-3 text-gray-400" />
                    )}
                  </span>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {modules?.length === 0 && (
        <div className="text-center py-12">
          <Puzzle className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-semibold text-gray-900">
            Sem módulos disponíveis
          </h3>
          <p className="mt-1 text-sm text-gray-500">
            Contacte o suporte para ativar módulos.
          </p>
        </div>
      )}
    </div>
  )
}
