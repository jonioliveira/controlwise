'use client'

import { useEffect } from 'react'
import Link from 'next/link'
import {
  Users,
  FileText,
  DollarSign,
  TrendingUp,
  Calendar,
  UserPlus,
  Stethoscope,
  FolderOpen,
  CheckSquare,
  Puzzle,
  ArrowRight,
  Loader2
} from 'lucide-react'
import { useEnabledModules } from '@/hooks/useModules'
import { useClients } from '@/hooks/useClients'
import { useAuthStore } from '@/stores/authStore'

export default function DashboardPage() {
  const { user } = useAuthStore()
  const { data: enabledModules = [], isLoading: modulesLoading, error: modulesError } = useEnabledModules()
  const { data: clientsData, isLoading: clientsLoading } = useClients({ limit: 5 })

  // Debug logging
  useEffect(() => {
    console.log('Dashboard Debug:', {
      user,
      enabledModules,
      modulesError,
    })
  }, [user, enabledModules, modulesError])

  const hasConstruction = enabledModules.includes('construction')
  const hasAppointments = enabledModules.includes('appointments')
  const hasNoModules = !hasConstruction && !hasAppointments

  if (modulesLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
      </div>
    )
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
      </div>

      {/* Core Stats - Always visible */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {/* Clients - Core feature */}
        <div className="card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600">Clientes</p>
              <p className="text-3xl font-bold text-gray-900 mt-2">
                {clientsLoading ? '...' : clientsData?.pagination?.total || 0}
              </p>
              <p className="text-sm mt-2 text-gray-500">total registados</p>
            </div>
            <div className="p-3 bg-primary-50 rounded-lg">
              <Users className="h-6 w-6 text-primary-600" />
            </div>
          </div>
        </div>

        {/* Construction Module Stats */}
        {hasConstruction && (
          <>
            <div className="card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">Orçamentos</p>
                  <p className="text-3xl font-bold text-gray-900 mt-2">-</p>
                  <p className="text-sm mt-2 text-gray-500">pendentes</p>
                </div>
                <div className="p-3 bg-amber-50 rounded-lg">
                  <FileText className="h-6 w-6 text-amber-600" />
                </div>
              </div>
            </div>

            <div className="card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">Projetos</p>
                  <p className="text-3xl font-bold text-gray-900 mt-2">-</p>
                  <p className="text-sm mt-2 text-gray-500">em curso</p>
                </div>
                <div className="p-3 bg-blue-50 rounded-lg">
                  <FolderOpen className="h-6 w-6 text-blue-600" />
                </div>
              </div>
            </div>

            <div className="card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">Pagamentos</p>
                  <p className="text-3xl font-bold text-gray-900 mt-2">-</p>
                  <p className="text-sm mt-2 text-gray-500">pendentes</p>
                </div>
                <div className="p-3 bg-green-50 rounded-lg">
                  <DollarSign className="h-6 w-6 text-green-600" />
                </div>
              </div>
            </div>
          </>
        )}

        {/* Appointments Module Stats */}
        {hasAppointments && (
          <>
            <div className="card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">Sessões Hoje</p>
                  <p className="text-3xl font-bold text-gray-900 mt-2">-</p>
                  <p className="text-sm mt-2 text-gray-500">agendadas</p>
                </div>
                <div className="p-3 bg-purple-50 rounded-lg">
                  <Calendar className="h-6 w-6 text-purple-600" />
                </div>
              </div>
            </div>

            <div className="card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">Pacientes</p>
                  <p className="text-3xl font-bold text-gray-900 mt-2">-</p>
                  <p className="text-sm mt-2 text-gray-500">ativos</p>
                </div>
                <div className="p-3 bg-teal-50 rounded-lg">
                  <UserPlus className="h-6 w-6 text-teal-600" />
                </div>
              </div>
            </div>

            <div className="card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600">Terapeutas</p>
                  <p className="text-3xl font-bold text-gray-900 mt-2">-</p>
                  <p className="text-sm mt-2 text-gray-500">registados</p>
                </div>
                <div className="p-3 bg-indigo-50 rounded-lg">
                  <Stethoscope className="h-6 w-6 text-indigo-600" />
                </div>
              </div>
            </div>
          </>
        )}
      </div>

      {/* No Modules Enabled */}
      {hasNoModules && (
        <div className="card mb-8">
          <div className="text-center py-8">
            <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
              <Puzzle className="h-8 w-8 text-gray-400" />
            </div>
            <h2 className="text-lg font-semibold text-gray-900 mb-2">
              Nenhum módulo ativo
            </h2>
            <p className="text-gray-500 max-w-md mx-auto mb-6">
              Ative módulos para desbloquear funcionalidades adicionais como gestão de obras,
              agendamentos, ou notificações.
            </p>
            <Link
              href="/dashboard/settings/modules"
              className="inline-flex items-center px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
            >
              Ver Módulos Disponíveis
              <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
          </div>
        </div>
      )}

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Clients - Always visible */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Clientes Recentes</h2>
          {clientsLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
            </div>
          ) : clientsData?.data && clientsData.data.length > 0 ? (
            <div className="space-y-3">
              {clientsData.data.slice(0, 5).map((client) => (
                <div key={client.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div className="flex items-center">
                    <div className="w-8 h-8 bg-primary-100 rounded-full flex items-center justify-center">
                      <span className="text-primary-700 text-sm font-medium">
                        {client.name.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div className="ml-3">
                      <p className="text-sm font-medium text-gray-900">{client.name}</p>
                      <p className="text-xs text-gray-500">{client.email}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <Users className="h-8 w-8 mx-auto mb-2 text-gray-300" />
              <p>Nenhum cliente registado</p>
            </div>
          )}
          <Link
            href="/dashboard/clients"
            className="mt-4 text-primary-600 hover:text-primary-700 text-sm font-medium inline-flex items-center"
          >
            Ver todos os clientes
            <ArrowRight className="ml-1 h-4 w-4" />
          </Link>
        </div>

        {/* Construction: Recent Projects */}
        {hasConstruction && (
          <div className="card">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Projetos Ativos</h2>
            <div className="text-center py-8 text-gray-500">
              <FolderOpen className="h-8 w-8 mx-auto mb-2 text-gray-300" />
              <p>Nenhum projeto ativo</p>
            </div>
            <Link
              href="/dashboard/projects"
              className="mt-4 text-primary-600 hover:text-primary-700 text-sm font-medium inline-flex items-center"
            >
              Ver todos os projetos
              <ArrowRight className="ml-1 h-4 w-4" />
            </Link>
          </div>
        )}

        {/* Appointments: Today's Sessions */}
        {hasAppointments && (
          <div className="card">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Sessões de Hoje</h2>
            <div className="text-center py-8 text-gray-500">
              <Calendar className="h-8 w-8 mx-auto mb-2 text-gray-300" />
              <p>Nenhuma sessão agendada para hoje</p>
            </div>
            <Link
              href="/dashboard/agenda"
              className="mt-4 text-primary-600 hover:text-primary-700 text-sm font-medium inline-flex items-center"
            >
              Ver agenda completa
              <ArrowRight className="ml-1 h-4 w-4" />
            </Link>
          </div>
        )}
      </div>

      {/* Quick Actions - Module aware */}
      <div className="mt-6 card">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Ações Rápidas</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {/* Always available */}
          <Link
            href="/dashboard/clients/new"
            className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 text-center transition-colors"
          >
            <Users className="h-6 w-6 text-primary-600 mx-auto mb-2" />
            <span className="text-sm font-medium text-gray-700">Novo Cliente</span>
          </Link>

          {/* Construction module */}
          {hasConstruction && (
            <>
              <Link
                href="/dashboard/worksheets/new"
                className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 text-center transition-colors"
              >
                <FileText className="h-6 w-6 text-amber-600 mx-auto mb-2" />
                <span className="text-sm font-medium text-gray-700">Nova Folha de Obra</span>
              </Link>
              <Link
                href="/dashboard/budgets/new"
                className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 text-center transition-colors"
              >
                <DollarSign className="h-6 w-6 text-green-600 mx-auto mb-2" />
                <span className="text-sm font-medium text-gray-700">Novo Orçamento</span>
              </Link>
              <Link
                href="/dashboard/tasks"
                className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 text-center transition-colors"
              >
                <CheckSquare className="h-6 w-6 text-blue-600 mx-auto mb-2" />
                <span className="text-sm font-medium text-gray-700">Ver Tarefas</span>
              </Link>
            </>
          )}

          {/* Appointments module */}
          {hasAppointments && (
            <>
              <Link
                href="/dashboard/agenda"
                className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 text-center transition-colors"
              >
                <Calendar className="h-6 w-6 text-purple-600 mx-auto mb-2" />
                <span className="text-sm font-medium text-gray-700">Nova Sessão</span>
              </Link>
              <Link
                href="/dashboard/patients/new"
                className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 text-center transition-colors"
              >
                <UserPlus className="h-6 w-6 text-teal-600 mx-auto mb-2" />
                <span className="text-sm font-medium text-gray-700">Novo Paciente</span>
              </Link>
              <Link
                href="/dashboard/therapists"
                className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 text-center transition-colors"
              >
                <Stethoscope className="h-6 w-6 text-indigo-600 mx-auto mb-2" />
                <span className="text-sm font-medium text-gray-700">Terapeutas</span>
              </Link>
            </>
          )}

          {/* Settings - always available */}
          <Link
            href="/dashboard/settings"
            className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 text-center transition-colors"
          >
            <TrendingUp className="h-6 w-6 text-gray-600 mx-auto mb-2" />
            <span className="text-sm font-medium text-gray-700">Definições</span>
          </Link>
        </div>
      </div>
    </div>
  )
}
