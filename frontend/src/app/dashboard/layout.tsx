'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import {
  Building2,
  LayoutDashboard,
  Users,
  FileText,
  DollarSign,
  FolderOpen,
  CheckSquare,
  CreditCard,
  Bell,
  Settings,
  LogOut,
  Menu,
  X,
  Calendar,
  UserPlus,
  Stethoscope
} from 'lucide-react'
import { useState, useMemo } from 'react'
import { api } from '@/lib/api'
import { useEnabledModules } from '@/hooks/useModules'
import { useOrganization } from '@/hooks/useOrganization'
import type { ModuleName } from '@/types'

interface NavItem {
  name: string
  href: string
  icon: React.ComponentType<{ className?: string }>
  module?: ModuleName
  // Special visibility rules
  hideWhenOnlyModule?: ModuleName // Hide when ONLY this module is active (not combined with others)
}

const allNavigation: NavItem[] = [
  // Core features
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  // Clients: Hide only in pure healthcare mode (appointments without construction)
  // When construction is active, clients are needed for project management
  { name: 'Clientes', href: '/dashboard/clients', icon: Users, hideWhenOnlyModule: 'appointments' },

  // Construction module - industry-specific features
  { name: 'Folhas de Obra', href: '/dashboard/worksheets', icon: FileText, module: 'construction' },
  { name: 'Orçamentos', href: '/dashboard/budgets', icon: DollarSign, module: 'construction' },
  { name: 'Projetos', href: '/dashboard/projects', icon: FolderOpen, module: 'construction' },
  { name: 'Tarefas', href: '/dashboard/tasks', icon: CheckSquare, module: 'construction' },
  { name: 'Pagamentos', href: '/dashboard/payments', icon: CreditCard, module: 'construction' },

  // Appointments module - healthcare/therapy features
  { name: 'Agenda', href: '/dashboard/agenda', icon: Calendar, module: 'appointments' },
  { name: 'Pacientes', href: '/dashboard/patients', icon: UserPlus, module: 'appointments' },
  { name: 'Terapeutas', href: '/dashboard/therapists', icon: Stethoscope, module: 'appointments' },
  { name: 'Pagamentos', href: '/dashboard/session-payments', icon: CreditCard, module: 'appointments' },
]

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const pathname = usePathname()
  const router = useRouter()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const { data: enabledModules = [] } = useEnabledModules()
  const { data: organization } = useOrganization()

  // Filter navigation based on enabled modules
  const navigation = useMemo(() => {
    const hasConstruction = enabledModules.includes('construction')
    const hasAppointments = enabledModules.includes('appointments')

    return allNavigation.filter((item) => {
      // Hide item only when ONLY the specified module is active (pure mode)
      // e.g., hide Clients only in pure healthcare mode (appointments without construction)
      if (item.hideWhenOnlyModule) {
        const onlyThatModule = enabledModules.includes(item.hideWhenOnlyModule) &&
          enabledModules.length === 1
        // Also hide when appointments is active but construction is not
        // (allows healthcare-focused orgs to use Patients instead of Clients)
        if (item.hideWhenOnlyModule === 'appointments' && hasAppointments && !hasConstruction) {
          return false
        }
      }
      // Items without a module requirement are visible
      if (!item.module) return true
      // Otherwise, check if the module is enabled
      return enabledModules.includes(item.module)
    })
  }, [enabledModules])

  const handleLogout = async () => {
    try {
      await api.logout()
      router.push('/login')
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Sidebar for desktop */}
      <div className="hidden md:fixed md:inset-y-0 md:flex md:w-64 md:flex-col">
        <div className="flex flex-col flex-grow bg-white border-r border-gray-200 pt-5 overflow-y-auto">
          <div className="flex-shrink-0 px-4">
            <div className="flex items-center">
              <Building2 className="h-8 w-8 text-primary-600" />
              <span className="ml-2 text-xl font-bold text-gray-900">ControleWise</span>
            </div>
            {organization && (
              <p className="mt-2 text-sm text-gray-500 truncate">{organization.name}</p>
            )}
          </div>

          <div className="mt-6 flex-grow flex flex-col">
            <nav className="flex-1 px-2 space-y-1">
              {navigation.map((item) => {
                const isActive = pathname === item.href
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    className={`
                      group flex items-center px-2 py-2 text-sm font-medium rounded-md
                      ${isActive 
                        ? 'bg-primary-50 text-primary-700' 
                        : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                      }
                    `}
                  >
                    <item.icon
                      className={`
                        mr-3 h-5 w-5
                        ${isActive ? 'text-primary-600' : 'text-gray-400 group-hover:text-gray-500'}
                      `}
                    />
                    {item.name}
                  </Link>
                )
              })}
            </nav>
          </div>

          <div className="flex-shrink-0 flex border-t border-gray-200 p-4">
            <div className="flex flex-col w-full space-y-2">
              <Link
                href="/dashboard/settings"
                className="flex items-center text-sm text-gray-700 hover:text-gray-900"
              >
                <Settings className="h-5 w-5 mr-2" />
                Definições
              </Link>
              <button
                onClick={handleLogout}
                className="flex items-center text-sm text-red-600 hover:text-red-700"
              >
                <LogOut className="h-5 w-5 mr-2" />
                Sair
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Mobile menu */}
      <div className="md:hidden">
        <div className="fixed top-0 left-0 right-0 z-40 flex items-center justify-between bg-white border-b border-gray-200 px-4 py-3">
          <div className="flex items-center min-w-0">
            <Building2 className="h-7 w-7 text-primary-600 flex-shrink-0" />
            <div className="ml-2 min-w-0">
              <span className="text-lg font-bold text-gray-900">ControleWise</span>
              {organization && (
                <p className="text-xs text-gray-500 truncate">{organization.name}</p>
              )}
            </div>
          </div>
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="text-gray-500 hover:text-gray-700"
          >
            {mobileMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </button>
        </div>

        {mobileMenuOpen && (
          <div className="fixed inset-0 z-30 bg-gray-600 bg-opacity-75" onClick={() => setMobileMenuOpen(false)}>
            <div className="fixed inset-y-0 left-0 w-64 bg-white" onClick={(e) => e.stopPropagation()}>
              <nav className="mt-16 px-2 space-y-1">
                {navigation.map((item) => {
                  const isActive = pathname === item.href
                  return (
                    <Link
                      key={item.name}
                      href={item.href}
                      onClick={() => setMobileMenuOpen(false)}
                      className={`
                        group flex items-center px-2 py-2 text-sm font-medium rounded-md
                        ${isActive 
                          ? 'bg-primary-50 text-primary-700' 
                          : 'text-gray-700 hover:bg-gray-50'
                        }
                      `}
                    >
                      <item.icon className="mr-3 h-5 w-5" />
                      {item.name}
                    </Link>
                  )
                })}
              </nav>
            </div>
          </div>
        )}
      </div>

      {/* Main content */}
      <div className="md:pl-64 flex flex-col flex-1">
        <main className="flex-1">
          <div className="py-6 md:py-8 px-4 sm:px-6 md:px-8 mt-14 md:mt-0">
            {children}
          </div>
        </main>
      </div>
    </div>
  )
}
