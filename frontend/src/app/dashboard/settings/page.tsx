'use client'

import Link from 'next/link'
import {
  Settings,
  Building2,
  Users,
  Puzzle,
  Bell,
  GitBranch,
  ChevronRight
} from 'lucide-react'

const settingsCategories = [
  {
    name: 'Organização',
    description: 'Dados da empresa, logo e informações fiscais',
    href: '/dashboard/settings/organization',
    icon: Building2
  },
  {
    name: 'Utilizadores',
    description: 'Gerir utilizadores e permissões',
    href: '/dashboard/settings/users',
    icon: Users
  },
  {
    name: 'Módulos',
    description: 'Ativar ou desativar módulos da plataforma',
    href: '/dashboard/settings/modules',
    icon: Puzzle
  },
  {
    name: 'Notificações',
    description: 'Configurar WhatsApp e lembretes automáticos',
    href: '/dashboard/settings/notifications',
    icon: Bell
  },
  {
    name: 'Workflows',
    description: 'Configurar fluxos de trabalho automatizados',
    href: '/dashboard/settings/workflows',
    icon: GitBranch
  }
]

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Definições</h1>
        <p className="mt-1 text-sm text-gray-500">
          Gerir configurações da organização e da plataforma
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        {settingsCategories.map((category) => (
          <Link
            key={category.name}
            href={category.href}
            className="group relative flex items-start gap-4 rounded-lg border border-gray-200 bg-white p-5 shadow-sm transition-all hover:border-primary-300 hover:shadow-md"
          >
            <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary-600 group-hover:bg-primary-100">
              <category.icon className="h-6 w-6" />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base font-semibold text-gray-900 group-hover:text-primary-600">
                {category.name}
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                {category.description}
              </p>
            </div>
            <ChevronRight className="h-5 w-5 text-gray-400 group-hover:text-primary-500 self-center" />
          </Link>
        ))}
      </div>
    </div>
  )
}
