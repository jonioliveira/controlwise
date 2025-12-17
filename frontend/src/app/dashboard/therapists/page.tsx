'use client'

import { useState } from 'react'
import {
  Plus,
  Search,
  Stethoscope,
  Phone,
  Mail,
  Clock,
  Euro,
  MoreVertical,
  Edit,
  Trash2,
  Loader2,
  Users
} from 'lucide-react'
import { useTherapists, useDeleteTherapist, useTherapistStats } from '@/hooks/useTherapists'
import { TherapistFormModal } from '@/components/appointments/TherapistFormModal'
import type { Therapist } from '@/types'

export default function TherapistsPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingTherapistId, setEditingTherapistId] = useState<string | null>(null)
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null)

  const { data, isLoading, error } = useTherapists({ search, page, limit: 20 })
  const { data: stats } = useTherapistStats()
  const deleteTherapist = useDeleteTherapist()

  const handleEdit = (therapist: Therapist) => {
    setEditingTherapistId(therapist.id)
    setIsModalOpen(true)
    setActiveDropdown(null)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Tem certeza que deseja eliminar este terapeuta?')) return
    setDeletingId(id)
    try {
      await deleteTherapist.mutateAsync(id)
    } catch (error) {
      console.error('Failed to delete therapist:', error)
    } finally {
      setDeletingId(null)
      setActiveDropdown(null)
    }
  }

  const handleCloseModal = () => {
    setIsModalOpen(false)
    setEditingTherapistId(null)
  }

  const formatWorkingHours = (workingHours: Record<string, { start: string; end: string }>) => {
    const days = Object.keys(workingHours).filter((day) => workingHours[day])
    if (days.length === 0) return 'Não definido'
    if (days.length === 7) return 'Todos os dias'
    if (days.length === 5 && !workingHours.saturday && !workingHours.sunday) {
      return 'Seg-Sex'
    }
    return `${days.length} dias/semana`
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Terapeutas</h1>
          <p className="mt-1 text-sm text-gray-500">
            Gerir terapeutas e disponibilidade
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
        >
          <Plus className="h-4 w-4 mr-2" />
          Novo Terapeuta
        </button>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
                <Users className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Total</p>
                <p className="text-xl font-semibold text-gray-900">{stats.total}</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100">
                <Stethoscope className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Ativos</p>
                <p className="text-xl font-semibold text-gray-900">{stats.active}</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-100">
                <Clock className="h-5 w-5 text-purple-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Sessões Hoje</p>
                <p className="text-xl font-semibold text-gray-900">{stats.total_sessions_today}</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-orange-100">
                <Clock className="h-5 w-5 text-orange-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Sessões Semana</p>
                <p className="text-xl font-semibold text-gray-900">{stats.total_sessions_week}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Search */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value)
              setPage(1)
            }}
            placeholder="Pesquisar por nome, especialidade ou email..."
            className="block w-full pl-10 pr-4 py-2 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
          />
        </div>
      </div>

      {/* Therapists List */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
          </div>
        ) : error ? (
          <div className="text-center py-12 text-red-600">
            Erro ao carregar terapeutas
          </div>
        ) : !data?.therapists?.length ? (
          <div className="text-center py-12">
            <Stethoscope className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-semibold text-gray-900">Sem terapeutas</h3>
            <p className="mt-1 text-sm text-gray-500">
              Comece por adicionar um novo terapeuta.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {data?.therapists?.map((therapist) => (
              <div
                key={therapist.id}
                className="flex items-center justify-between p-4 hover:bg-gray-50"
              >
                <div className="flex items-center gap-4">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary-100 text-primary-600 font-medium">
                    {therapist.name.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="text-sm font-medium text-gray-900">
                        {therapist.name}
                      </h3>
                      {therapist.specialty && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800">
                          {therapist.specialty}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-4 mt-1">
                      {therapist.phone && (
                        <span className="flex items-center text-sm text-gray-500">
                          <Phone className="h-4 w-4 mr-1" />
                          {therapist.phone}
                        </span>
                      )}
                      {therapist.email && (
                        <span className="flex items-center text-sm text-gray-500">
                          <Mail className="h-4 w-4 mr-1" />
                          {therapist.email}
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <p className="text-sm text-gray-900 flex items-center justify-end">
                      <Clock className="h-4 w-4 mr-1 text-gray-400" />
                      {therapist.session_duration_minutes} min
                    </p>
                    <p className="text-sm text-gray-500 flex items-center justify-end">
                      <Euro className="h-4 w-4 mr-1" />
                      {(therapist.default_price_cents / 100).toFixed(2)}
                    </p>
                  </div>

                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      therapist.is_active
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {therapist.is_active ? 'Ativo' : 'Inativo'}
                  </span>

                  <div className="relative">
                    <button
                      onClick={() => setActiveDropdown(activeDropdown === therapist.id ? null : therapist.id)}
                      className="p-2 hover:bg-gray-100 rounded-md"
                    >
                      <MoreVertical className="h-5 w-5 text-gray-400" />
                    </button>
                    {activeDropdown === therapist.id && (
                      <div className="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 z-10">
                        <div className="py-1">
                          <button
                            onClick={() => handleEdit(therapist)}
                            className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                          >
                            <Edit className="h-4 w-4 mr-2" />
                            Editar
                          </button>
                          <button
                            onClick={() => handleDelete(therapist.id)}
                            disabled={deletingId === therapist.id}
                            className="flex items-center w-full px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                          >
                            {deletingId === therapist.id ? (
                              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                            ) : (
                              <Trash2 className="h-4 w-4 mr-2" />
                            )}
                            Eliminar
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {data && data.total > data.limit && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-200">
            <p className="text-sm text-gray-700">
              Mostrando {((page - 1) * data.limit) + 1} a {Math.min(page * data.limit, data.total)} de {data.total}
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
              >
                Anterior
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * data.limit >= data.total}
                className="px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
              >
                Seguinte
              </button>
            </div>
          </div>
        )}
      </div>

      <TherapistFormModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        therapistId={editingTherapistId}
      />
    </div>
  )
}
