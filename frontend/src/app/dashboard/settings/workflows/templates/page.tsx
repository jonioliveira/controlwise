'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ArrowLeft, Plus, Mail, MessageCircle, Edit2, Trash2, Loader2, Eye } from 'lucide-react'
import {
  useTemplates,
  useCreateTemplate,
  useUpdateTemplate,
  useDeleteTemplate,
} from '@/hooks/useWorkflows'
import { Modal } from '@/components/ui/Modal'
import { useForm } from 'react-hook-form'
import type { MessageTemplate } from '@/types'

const channelLabels: Record<string, string> = {
  email: 'Email',
  whatsapp: 'WhatsApp',
}

const channelIcons: Record<string, React.ReactNode> = {
  email: <Mail className="h-4 w-4" />,
  whatsapp: <MessageCircle className="h-4 w-4" />,
}

export default function TemplatesPage() {
  const [selectedChannel, setSelectedChannel] = useState<string>('')
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<MessageTemplate | null>(null)
  const [previewTemplate, setPreviewTemplate] = useState<MessageTemplate | null>(null)

  const { data: templates = [], isLoading } = useTemplates(selectedChannel || undefined)
  const deleteTemplate = useDeleteTemplate()

  const handleDelete = (id: string, name: string) => {
    if (confirm(`Tem certeza que deseja eliminar o modelo "${name}"? Esta ação não pode ser desfeita.`)) {
      deleteTemplate.mutate(id)
    }
  }

  const groupedTemplates = templates.reduce((acc, template) => {
    const channel = template.channel
    if (!acc[channel]) {
      acc[channel] = []
    }
    acc[channel].push(template)
    return acc
  }, {} as Record<string, MessageTemplate[]>)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link
            href="/dashboard/settings/workflows"
            className="p-2 text-gray-400 hover:text-gray-600 rounded-md hover:bg-gray-100"
          >
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Modelos de Mensagem</h1>
            <p className="text-sm text-gray-500">
              Configure modelos de email e WhatsApp para notificações automáticas
            </p>
          </div>
        </div>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
        >
          <Plus className="h-4 w-4" />
          Novo Modelo
        </button>
      </div>

      {/* Channel Filter */}
      <div className="flex items-center gap-2">
        <button
          onClick={() => setSelectedChannel('')}
          className={`px-3 py-1.5 text-sm rounded-md ${
            selectedChannel === ''
              ? 'bg-primary-100 text-primary-700'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          }`}
        >
          Todos
        </button>
        <button
          onClick={() => setSelectedChannel('email')}
          className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md ${
            selectedChannel === 'email'
              ? 'bg-primary-100 text-primary-700'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          }`}
        >
          <Mail className="h-4 w-4" />
          Email
        </button>
        <button
          onClick={() => setSelectedChannel('whatsapp')}
          className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md ${
            selectedChannel === 'whatsapp'
              ? 'bg-primary-100 text-primary-700'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          }`}
        >
          <MessageCircle className="h-4 w-4" />
          WhatsApp
        </button>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
        </div>
      ) : templates.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <Mail className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900">Nenhum modelo</h3>
          <p className="mt-1 text-sm text-gray-500">
            Crie modelos de mensagem para usar nas notificações automáticas
          </p>
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
          >
            <Plus className="h-4 w-4" />
            Criar Modelo
          </button>
        </div>
      ) : (
        <div className="space-y-8">
          {Object.entries(groupedTemplates).map(([channel, channelTemplates]) => (
            <div key={channel}>
              <h2 className="text-lg font-medium text-gray-900 mb-4 flex items-center gap-2">
                {channelIcons[channel]}
                {channelLabels[channel] || channel}
              </h2>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {channelTemplates.map((template) => (
                  <TemplateCard
                    key={template.id}
                    template={template}
                    onEdit={() => setEditingTemplate(template)}
                    onPreview={() => setPreviewTemplate(template)}
                    onDelete={() => handleDelete(template.id, template.name)}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create/Edit Modal */}
      {(isCreateModalOpen || editingTemplate) && (
        <TemplateFormModal
          template={editingTemplate}
          onClose={() => {
            setIsCreateModalOpen(false)
            setEditingTemplate(null)
          }}
        />
      )}

      {/* Preview Modal */}
      {previewTemplate && (
        <TemplatePreviewModal
          template={previewTemplate}
          onClose={() => setPreviewTemplate(null)}
        />
      )}
    </div>
  )
}

function TemplateCard({
  template,
  onEdit,
  onPreview,
  onDelete,
}: {
  template: MessageTemplate
  onEdit: () => void
  onPreview: () => void
  onDelete: () => void
}) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            {template.channel === 'email' ? (
              <Mail className="h-4 w-4 text-blue-500" />
            ) : (
              <MessageCircle className="h-4 w-4 text-green-500" />
            )}
            <h3 className="text-base font-medium text-gray-900 truncate">
              {template.name}
            </h3>
          </div>
          {template.subject && (
            <p className="text-sm text-gray-500 mt-1 truncate">
              Assunto: {template.subject}
            </p>
          )}
        </div>
        <div className={`px-2 py-0.5 text-xs rounded-full ${
          template.is_active
            ? 'bg-green-100 text-green-700'
            : 'bg-gray-100 text-gray-500'
        }`}>
          {template.is_active ? 'Ativo' : 'Inativo'}
        </div>
      </div>

      <p className="mt-3 text-sm text-gray-600 line-clamp-3 whitespace-pre-wrap">
        {template.body}
      </p>

      <div className="mt-4 flex items-center gap-2 pt-3 border-t border-gray-100">
        <button
          onClick={onPreview}
          className="flex-1 inline-flex items-center justify-center gap-1 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50 rounded-md"
        >
          <Eye className="h-4 w-4" />
          Pré-visualizar
        </button>
        <button
          onClick={onEdit}
          className="p-1.5 text-gray-400 hover:text-primary-600 rounded-md"
          title="Editar"
        >
          <Edit2 className="h-4 w-4" />
        </button>
        <button
          onClick={onDelete}
          className="p-1.5 text-gray-400 hover:text-red-600 rounded-md"
          title="Eliminar"
        >
          <Trash2 className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}

interface TemplateFormData {
  name: string
  channel: 'email' | 'whatsapp'
  subject?: string
  body: string
  is_active: boolean
}

function TemplateFormModal({
  template,
  onClose,
}: {
  template: MessageTemplate | null
  onClose: () => void
}) {
  const createTemplate = useCreateTemplate()
  const updateTemplate = useUpdateTemplate()
  const isEditing = !!template

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<TemplateFormData>({
    defaultValues: template
      ? {
          name: template.name,
          channel: template.channel,
          subject: template.subject || '',
          body: template.body,
          is_active: template.is_active,
        }
      : {
          name: '',
          channel: 'email',
          subject: '',
          body: '',
          is_active: true,
        },
  })

  const channel = watch('channel')

  const onSubmit = async (data: TemplateFormData) => {
    try {
      if (isEditing && template) {
        await updateTemplate.mutateAsync({
          id: template.id,
          data: {
            name: data.name,
            channel: data.channel,
            subject: data.channel === 'email' ? data.subject : undefined,
            body: data.body,
            is_active: data.is_active,
          },
        })
      } else {
        await createTemplate.mutateAsync({
          name: data.name,
          channel: data.channel,
          subject: data.channel === 'email' ? data.subject : undefined,
          body: data.body,
        })
      }
      onClose()
    } catch (error) {
      console.error('Failed to save template:', error)
    }
  }

  return (
    <Modal
      isOpen={true}
      onClose={onClose}
      title={isEditing ? 'Editar Modelo' : 'Novo Modelo'}
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Nome *
          </label>
          <input
            type="text"
            {...register('name', { required: 'Nome é obrigatório' })}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            placeholder="Ex: Lembrete de Consulta"
          />
          {errors.name && (
            <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Canal *
          </label>
          <select
            {...register('channel')}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
          >
            <option value="email">Email</option>
            <option value="whatsapp">WhatsApp</option>
          </select>
        </div>

        {channel === 'email' && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Assunto
            </label>
            <input
              type="text"
              {...register('subject')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="Ex: Lembrete: Consulta amanhã"
            />
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Corpo da Mensagem *
          </label>
          <textarea
            {...register('body', { required: 'Corpo é obrigatório' })}
            rows={8}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 font-mono text-sm"
            placeholder="Use {{variavel}} para campos dinâmicos"
          />
          {errors.body && (
            <p className="mt-1 text-sm text-red-600">{errors.body.message}</p>
          )}
          <p className="mt-1 text-xs text-gray-500">
            Variáveis disponíveis: {`{{patient_name}}, {{session_date}}, {{session_time}}, {{therapist_name}}, {{client_name}}, {{project_name}}, {{budget_total}}, {{organization_name}}`}
          </p>
        </div>

        {isEditing && (
          <div className="flex items-center">
            <input
              type="checkbox"
              {...register('is_active')}
              id="is_active"
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <label htmlFor="is_active" className="ml-2 text-sm text-gray-700">
              Modelo ativo
            </label>
          </div>
        )}

        <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 disabled:opacity-50"
          >
            {isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
            {isEditing ? 'Guardar' : 'Criar'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

function TemplatePreviewModal({
  template,
  onClose,
}: {
  template: MessageTemplate
  onClose: () => void
}) {
  // Sample data for preview
  const sampleData: Record<string, string> = {
    patient_name: 'João Silva',
    patient_phone: '+351912345678',
    patient_email: 'joao.silva@email.com',
    therapist_name: 'Dr. Maria Santos',
    session_date: '15/01/2025',
    session_time: '14:30',
    session_type: 'Consulta Regular',
    amount: '50.00',
    client_name: 'Manuel Costa',
    client_email: 'manuel.costa@email.com',
    project_name: 'Remodelação Cozinha',
    budget_total: '15000.00',
    budget_number: 'ORC-2025-001',
    budget_link: 'https://app.controlewise.pt/budgets/123',
    organization_name: 'Clínica Exemplo',
  }

  const renderTemplate = (text: string) => {
    return text.replace(/\{\{(\w+)\}\}/g, (match, varName) => {
      return sampleData[varName] || match
    })
  }

  const renderedBody = renderTemplate(template.body)
  const renderedSubject = template.subject ? renderTemplate(template.subject) : null

  return (
    <Modal isOpen={true} onClose={onClose} title="Pré-visualização">
      <div className="space-y-4">
        <div className="flex items-center gap-2 text-sm text-gray-500">
          {template.channel === 'email' ? (
            <>
              <Mail className="h-4 w-4" />
              Email
            </>
          ) : (
            <>
              <MessageCircle className="h-4 w-4" />
              WhatsApp
            </>
          )}
        </div>

        {renderedSubject && (
          <div>
            <label className="block text-xs font-medium text-gray-500 mb-1">
              Assunto
            </label>
            <div className="p-3 bg-gray-50 rounded-md text-sm">
              {renderedSubject}
            </div>
          </div>
        )}

        <div>
          <label className="block text-xs font-medium text-gray-500 mb-1">
            Mensagem
          </label>
          <div className="p-4 bg-gray-50 rounded-md whitespace-pre-wrap text-sm">
            {renderedBody}
          </div>
        </div>

        <div className="pt-4 border-t border-gray-200">
          <p className="text-xs text-gray-500 mb-2">
            Dados de exemplo usados na pré-visualização:
          </p>
          <div className="grid grid-cols-2 gap-1 text-xs text-gray-500">
            {Object.entries(sampleData).slice(0, 6).map(([key, value]) => (
              <div key={key}>
                <span className="font-mono">{`{{${key}}}`}</span> = {value}
              </div>
            ))}
          </div>
        </div>

        <div className="flex justify-end pt-4">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Fechar
          </button>
        </div>
      </div>
    </Modal>
  )
}
