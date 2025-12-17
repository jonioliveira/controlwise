'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, FileText, Edit2, Trash2, Eye } from 'lucide-react'
import { api } from '@/lib/api'
import { Modal } from '@/components/ui/Modal'
import { StatusBadge } from '@/components/ui/StatusBadge'
import type { WorkSheet, WorkSheetItem, Client } from '@/types'
import { format } from 'date-fns'

export default function WorksheetsPage() {
  const queryClient = useQueryClient()
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [viewingWorksheet, setViewingWorksheet] = useState<string | null>(null)
  const [editingWorksheet, setEditingWorksheet] = useState<WorkSheet | null>(null)

  // Fetch worksheets
  const { data: worksheets, isLoading } = useQuery({
    queryKey: ['worksheets'],
    queryFn: () => api.getWorksheets(),
  })

  // Fetch clients for the dropdown
  const { data: clients } = useQuery({
    queryKey: ['clients'],
    queryFn: () => api.getClients(),
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteWorksheet(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['worksheets'] })
    },
  })

  const handleView = (id: string) => {
    setViewingWorksheet(id)
  }

  const handleEdit = (worksheet: WorkSheet) => {
    setEditingWorksheet(worksheet)
    setIsModalOpen(true)
  }

  const handleDelete = async (id: string) => {
    if (confirm('Tem a certeza que deseja eliminar esta folha de obra?')) {
      await deleteMutation.mutateAsync(id)
    }
  }

  const handleCloseModal = () => {
    setIsModalOpen(false)
    setEditingWorksheet(null)
  }

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
        <p className="mt-4 text-gray-600">A carregar folhas de obra...</p>
      </div>
    )
  }

  return (
    <div>
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Folhas de Obra</h1>
          <p className="text-gray-600 mt-1">Gerir folhas de obra dos projetos</p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="btn btn-primary flex items-center"
        >
          <Plus className="h-5 w-5 mr-2" />
          Nova Folha de Obra
        </button>
      </div>

      {/* Worksheets List */}
      {!worksheets || worksheets.length === 0 ? (
        <div className="card text-center py-12">
          <FileText className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-600 mb-4">Ainda não tem folhas de obra cadastradas.</p>
          <button onClick={() => setIsModalOpen(true)} className="btn btn-primary">
            Criar Primeira Folha de Obra
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6">
          {worksheets.map((worksheet: any) => (
            <div key={worksheet.id} className="card hover:shadow-lg transition-shadow">
              <div className="flex justify-between items-start mb-4">
                <div className="flex-1">
                  <div className="flex items-center space-x-3 mb-2">
                    <h3 className="text-lg font-semibold text-gray-900">
                      {worksheet.title}
                    </h3>
                    <StatusBadge status={worksheet.status} />
                  </div>
                  <p className="text-sm text-gray-600 mb-2">
                    Cliente: <span className="font-medium">{worksheet.client_name}</span>
                  </p>
                  <p className="text-gray-700 line-clamp-2">{worksheet.description}</p>
                </div>
                <div className="flex space-x-2 ml-4">
                  <button
                    onClick={() => handleView(worksheet.id)}
                    className="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
                    title="Ver detalhes"
                  >
                    <Eye className="h-4 w-4" />
                  </button>
                  {worksheet.status === 'draft' && (
                    <>
                      <button
                        onClick={() => handleEdit(worksheet)}
                        className="p-2 text-primary-600 hover:bg-primary-50 rounded-lg transition-colors"
                        title="Editar"
                      >
                        <Edit2 className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => handleDelete(worksheet.id)}
                        className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                        title="Eliminar"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </>
                  )}
                </div>
              </div>

              <div className="flex items-center justify-between pt-4 border-t border-gray-200 text-sm text-gray-500">
                <span>
                  {worksheet.items?.length || 0} {worksheet.items?.length === 1 ? 'item' : 'items'}
                </span>
                <span>
                  Criado em {format(new Date(worksheet.created_at), 'dd/MM/yyyy')}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create/Edit Modal */}
      {isModalOpen && (
        <WorksheetFormModal
          worksheet={editingWorksheet}
          clients={clients || []}
          onClose={handleCloseModal}
          onSuccess={() => {
            queryClient.invalidateQueries({ queryKey: ['worksheets'] })
            handleCloseModal()
          }}
        />
      )}

      {/* View Modal */}
      {viewingWorksheet && (
        <WorksheetViewModal
          worksheetId={viewingWorksheet}
          onClose={() => setViewingWorksheet(null)}
        />
      )}
    </div>
  )
}

// Worksheet Form Modal
function WorksheetFormModal({
  worksheet,
  clients,
  onClose,
  onSuccess,
}: {
  worksheet: WorkSheet | null
  clients: Client[]
  onClose: () => void
  onSuccess: () => void
}) {
  const [formData, setFormData] = useState({
    client_id: worksheet?.client_id || '',
    title: worksheet?.title || '',
    description: worksheet?.description || '',
  })
  const [items, setItems] = useState<Partial<WorkSheetItem>[]>(
    worksheet ? [] : [{ description: '', quantity: 1, unit: 'm²', notes: '' }]
  )
  const [error, setError] = useState('')

  const mutation = useMutation({
    mutationFn: (data: any) => {
      if (worksheet) {
        return api.updateWorksheet(worksheet.id, data)
      }
      return api.createWorksheet(data)
    },
    onSuccess: () => {
      onSuccess()
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || 'Erro ao guardar folha de obra')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!formData.client_id || !formData.title || !formData.description) {
      setError('Todos os campos são obrigatórios')
      return
    }

    if (items.length === 0 || items.some((i) => !i.description)) {
      setError('Adicione pelo menos um item válido')
      return
    }

    mutation.mutate({ ...formData, items })
  }

  const addItem = () => {
    setItems([...items, { description: '', quantity: 1, unit: 'm²', notes: '' }])
  }

  const removeItem = (index: number) => {
    setItems(items.filter((_, i) => i !== index))
  }

  const updateItem = (index: number, field: string, value: any) => {
    const newItems = [...items]
    newItems[index] = { ...newItems[index], [field]: value }
    setItems(newItems)
  }

  return (
    <Modal
      isOpen={true}
      onClose={onClose}
      title={worksheet ? 'Editar Folha de Obra' : 'Nova Folha de Obra'}
      size="lg"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
            {error}
          </div>
        )}

        <div>
          <label className="label">Cliente *</label>
          <select
            required
            className="input"
            value={formData.client_id}
            onChange={(e) => setFormData({ ...formData, client_id: e.target.value })}
          >
            <option value="">Selecionar cliente</option>
            {clients.map((client) => (
              <option key={client.id} value={client.id}>
                {client.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="label">Título *</label>
          <input
            type="text"
            required
            className="input"
            value={formData.title}
            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
          />
        </div>

        <div>
          <label className="label">Descrição *</label>
          <textarea
            required
            rows={3}
            className="input"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
          />
        </div>

        {/* Items */}
        <div>
          <div className="flex justify-between items-center mb-2">
            <label className="label mb-0">Items *</label>
            <button type="button" onClick={addItem} className="text-sm text-primary-600 hover:text-primary-700">
              + Adicionar Item
            </button>
          </div>
          <div className="space-y-3">
            {items.map((item, index) => (
              <div key={index} className="p-4 border border-gray-200 rounded-lg space-y-3">
                <div className="flex justify-between items-start">
                  <span className="text-sm font-medium text-gray-700">Item {index + 1}</span>
                  {items.length > 1 && (
                    <button
                      type="button"
                      onClick={() => removeItem(index)}
                      className="text-red-600 hover:text-red-700"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  )}
                </div>
                <input
                  type="text"
                  placeholder="Descrição"
                  required
                  className="input"
                  value={item.description}
                  onChange={(e) => updateItem(index, 'description', e.target.value)}
                />
                <div className="grid grid-cols-2 gap-3">
                  <input
                    type="number"
                    placeholder="Quantidade"
                    required
                    min="0"
                    step="0.01"
                    className="input"
                    value={item.quantity}
                    onChange={(e) => updateItem(index, 'quantity', parseFloat(e.target.value))}
                  />
                  <input
                    type="text"
                    placeholder="Unidade (m², m³, un)"
                    required
                    className="input"
                    value={item.unit}
                    onChange={(e) => updateItem(index, 'unit', e.target.value)}
                  />
                </div>
                <input
                  type="text"
                  placeholder="Notas (opcional)"
                  className="input"
                  value={item.notes}
                  onChange={(e) => updateItem(index, 'notes', e.target.value)}
                />
              </div>
            ))}
          </div>
        </div>

        <div className="flex justify-end space-x-3 pt-4">
          <button type="button" onClick={onClose} className="btn btn-secondary">
            Cancelar
          </button>
          <button type="submit" disabled={mutation.isPending} className="btn btn-primary">
            {mutation.isPending ? 'A guardar...' : 'Guardar'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

// View Modal (simplified - you can expand this)
function WorksheetViewModal({
  worksheetId,
  onClose,
}: {
  worksheetId: string
  onClose: () => void
}) {
  // TODO: Fetch worksheet details with items
  return (
    <Modal isOpen={true} onClose={onClose} title="Detalhes da Folha de Obra" size="lg">
      <div className="text-center py-8">
        <p className="text-gray-600">Visualização detalhada em desenvolvimento...</p>
      </div>
    </Modal>
  )
}
