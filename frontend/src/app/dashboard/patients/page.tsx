'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import {
  Plus,
  Search,
  UserPlus,
  Phone,
  Mail,
  Calendar,
  MoreVertical,
  Edit,
  Trash2,
  Loader2,
  Users
} from 'lucide-react'
import { usePatients, useDeletePatient, usePatientStats } from '@/hooks/usePatients'
import { PatientFormModal } from '@/components/appointments/PatientFormModal'
import type { Patient } from '@/types'

export default function PatientsPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingPatientId, setEditingPatientId] = useState<string | null>(null)
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null)

  const { data, isLoading, error } = usePatients({ search, page, limit: 20 })
  const { data: stats } = usePatientStats()
  const deletePatient = useDeletePatient()

  const handleEdit = (patient: Patient) => {
    setEditingPatientId(patient.id)
    setIsModalOpen(true)
    setActiveDropdown(null)
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Tem certeza que deseja eliminar este paciente?')) return
    setDeletingId(id)
    try {
      await deletePatient.mutateAsync(id)
    } catch (error) {
      console.error('Failed to delete patient:', error)
    } finally {
      setDeletingId(null)
      setActiveDropdown(null)
    }
  }

  const handleCloseModal = () => {
    setIsModalOpen(false)
    setEditingPatientId(null)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Pacientes</h1>
          <p className="mt-1 text-sm text-gray-500">
            Gerir pacientes e informações de contacto
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
        >
          <Plus className="h-4 w-4 mr-2" />
          Novo Paciente
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
                <UserPlus className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Ativos</p>
                <p className="text-xl font-semibold text-gray-900">{stats.active}</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-100">
                <Users className="h-5 w-5 text-gray-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Inativos</p>
                <p className="text-xl font-semibold text-gray-900">{stats.inactive}</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple-100">
                <Calendar className="h-5 w-5 text-purple-600" />
              </div>
              <div>
                <p className="text-sm text-gray-500">Novos (mês)</p>
                <p className="text-xl font-semibold text-gray-900">{stats.new_this_month}</p>
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
            placeholder="Pesquisar por nome, telefone ou email..."
            className="block w-full pl-10 pr-4 py-2 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
          />
        </div>
      </div>

      {/* Patients List */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
          </div>
        ) : error ? (
          <div className="text-center py-12 text-red-600">
            Erro ao carregar pacientes
          </div>
        ) : !data?.patients?.length ? (
          <div className="text-center py-12">
            <UserPlus className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-semibold text-gray-900">Sem pacientes</h3>
            <p className="mt-1 text-sm text-gray-500">
              Comece por adicionar um novo paciente.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {data?.patients?.map((patient) => (
              <div
                key={patient.id}
                className="flex items-center justify-between p-4 hover:bg-gray-50"
              >
                <div className="flex items-center gap-4">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary-100 text-primary-600 font-medium">
                    {patient.client_name.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <h3 className="text-sm font-medium text-gray-900">
                      {patient.client_name}
                    </h3>
                    <div className="flex items-center gap-4 mt-1">
                      <span className="flex items-center text-sm text-gray-500">
                        <Phone className="h-4 w-4 mr-1" />
                        {patient.client_phone}
                      </span>
                      {patient.client_email && (
                        <span className="flex items-center text-sm text-gray-500">
                          <Mail className="h-4 w-4 mr-1" />
                          {patient.client_email}
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      patient.is_active
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {patient.is_active ? 'Ativo' : 'Inativo'}
                  </span>

                  <div className="relative">
                    <button
                      onClick={() => setActiveDropdown(activeDropdown === patient.id ? null : patient.id)}
                      className="p-2 hover:bg-gray-100 rounded-md"
                    >
                      <MoreVertical className="h-5 w-5 text-gray-400" />
                    </button>
                    {activeDropdown === patient.id && (
                      <div className="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 z-10">
                        <div className="py-1">
                          <button
                            onClick={() => handleEdit(patient)}
                            className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                          >
                            <Edit className="h-4 w-4 mr-2" />
                            Editar
                          </button>
                          <button
                            onClick={() => handleDelete(patient.id)}
                            disabled={deletingId === patient.id}
                            className="flex items-center w-full px-4 py-2 text-sm text-red-600 hover:bg-red-50"
                          >
                            {deletingId === patient.id ? (
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

      <PatientFormModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        patientId={editingPatientId}
      />
    </div>
  )
}
