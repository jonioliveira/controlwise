'use client'

import { useState, useMemo } from 'react'
import { Plus, Search, Edit2, Trash2, Mail, Phone, MapPin } from 'lucide-react'
import { useClients, useDeleteClient } from '@/hooks/useClients'
import { ClientFormModal } from '@/components/clients/ClientFormModal'
import type { Client } from '@/types'

export default function ClientsPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingClient, setEditingClient] = useState<Client | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)

  const { data: clientsData, isLoading } = useClients()
  const deleteMutation = useDeleteClient()

  const clients = clientsData?.data || []

  const filteredClients = useMemo(() => {
    if (!clients || !Array.isArray(clients)) return []
    if (!searchQuery) return clients

    const query = searchQuery.toLowerCase()
    return clients.filter(
      (client) =>
        client.name.toLowerCase().includes(query) ||
        client.email.toLowerCase().includes(query) ||
        client.phone?.includes(query)
    )
  }, [clients, searchQuery])

  const handleEdit = (client: Client) => {
    setEditingClient(client)
    setIsModalOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (deleteConfirm === id) {
      await deleteMutation.mutateAsync(id)
      setDeleteConfirm(null)
    } else {
      setDeleteConfirm(id)
      setTimeout(() => setDeleteConfirm(null), 3000)
    }
  }

  const handleCloseModal = () => {
    setIsModalOpen(false)
    setEditingClient(null)
  }

  return (
    <div>
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Clientes</h1>
          <p className="text-gray-600 mt-1">Gerir clientes da empresa</p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="btn btn-primary flex items-center"
        >
          <Plus className="h-5 w-5 mr-2" />
          Novo Cliente
        </button>
      </div>

      {/* Search */}
      <div className="card mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
          <input
            type="text"
            placeholder="Pesquisar por nome ou email..."
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Clients Grid */}
      {isLoading ? (
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">A carregar clientes...</p>
        </div>
      ) : filteredClients.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-600">
            {searchQuery ? 'Nenhum cliente encontrado.' : 'Ainda não tem clientes cadastrados.'}
          </p>
          {!searchQuery && (
            <button
              onClick={() => setIsModalOpen(true)}
              className="mt-4 btn btn-primary"
            >
              Adicionar Primeiro Cliente
            </button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredClients.map((client) => (
            <ClientCard
              key={client.id}
              client={client}
              onEdit={handleEdit}
              onDelete={handleDelete}
              isDeleteConfirm={deleteConfirm === client.id}
            />
          ))}
        </div>
      )}

      {/* Modal */}
      <ClientFormModal
        client={editingClient}
        isOpen={isModalOpen}
        onClose={handleCloseModal}
      />
    </div>
  )
}

interface ClientCardProps {
  client: Client
  onEdit: (client: Client) => void
  onDelete: (id: string) => void
  isDeleteConfirm: boolean
}

function ClientCard({ client, onEdit, onDelete, isDeleteConfirm }: ClientCardProps) {
  return (
    <div className="card hover:shadow-lg transition-shadow">
      <div className="flex justify-between items-start mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">{client.name}</h3>
          {client.tax_id && (
            <p className="text-sm text-gray-500">NIF: {client.tax_id}</p>
          )}
        </div>
        <div className="flex space-x-2">
          <button
            onClick={() => onEdit(client)}
            className="p-2 text-primary-600 hover:bg-primary-50 rounded-lg transition-colors"
            aria-label="Editar cliente"
          >
            <Edit2 className="h-4 w-4" />
          </button>
          <button
            onClick={() => onDelete(client.id)}
            className={`p-2 rounded-lg transition-colors ${
              isDeleteConfirm
                ? 'bg-red-100 text-red-700'
                : 'text-red-600 hover:bg-red-50'
            }`}
            aria-label={isDeleteConfirm ? 'Confirmar eliminação' : 'Eliminar cliente'}
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div className="space-y-2 text-sm">
        <div className="flex items-center text-gray-600">
          <Mail className="h-4 w-4 mr-2 flex-shrink-0" />
          <a
            href={`mailto:${client.email}`}
            className="hover:text-primary-600 truncate"
          >
            {client.email}
          </a>
        </div>
        <div className="flex items-center text-gray-600">
          <Phone className="h-4 w-4 mr-2 flex-shrink-0" />
          <a href={`tel:${client.phone}`} className="hover:text-primary-600">
            {client.phone}
          </a>
        </div>
        {client.address && (
          <div className="flex items-start text-gray-600">
            <MapPin className="h-4 w-4 mr-2 flex-shrink-0 mt-0.5" />
            <span className="line-clamp-2">{client.address}</span>
          </div>
        )}
      </div>

      {client.notes && (
        <div className="mt-4 pt-4 border-t border-gray-200">
          <p className="text-sm text-gray-600 line-clamp-2">{client.notes}</p>
        </div>
      )}

      {isDeleteConfirm && (
        <div className="mt-4 p-2 bg-red-50 border border-red-200 rounded text-sm text-red-700 text-center">
          Clique novamente para confirmar
        </div>
      )}
    </div>
  )
}
