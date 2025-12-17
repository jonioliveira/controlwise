'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, Settings, Copy, Trash2, CheckCircle, XCircle, Loader2, Wand2, Activity, Mail } from 'lucide-react'
import Link from 'next/link'
import {
  useWorkflows,
  useCreateWorkflow,
  useDeleteWorkflow,
  useDuplicateWorkflow,
  useInitDefaultWorkflows,
} from '@/hooks/useWorkflows'
import { useEnabledModules } from '@/hooks/useModules'
import { Modal } from '@/components/ui/Modal'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { workflowSchema, type WorkflowFormData } from '@/schemas/workflow'
import type { WorkflowWithStats } from '@/types'

const moduleLabels: Record<string, string> = {
  appointments: 'Consultas',
  construction: 'Construção',
}

const entityTypeLabels: Record<string, string> = {
  session: 'Sessão',
  budget: 'Orçamento',
  project: 'Projeto',
}

export default function WorkflowsPage() {
  const router = useRouter()
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [duplicatingWorkflow, setDuplicatingWorkflow] = useState<WorkflowWithStats | null>(null)

  const { data: workflows = [], isLoading } = useWorkflows()
  const { data: enabledModules = [] } = useEnabledModules()
  const deleteWorkflow = useDeleteWorkflow()
  const initDefaultWorkflows = useInitDefaultWorkflows()

  const handleInitDefaults = () => {
    initDefaultWorkflows.mutate('construction')
  }

  const handleEdit = (workflow: WorkflowWithStats) => {
    router.push(`/dashboard/settings/workflows/${workflow.id}`)
  }

  const handleDelete = (id: string) => {
    if (confirm('Tem certeza que deseja eliminar este workflow? Esta ação não pode ser desfeita.')) {
      deleteWorkflow.mutate(id)
    }
  }

  const groupedWorkflows = workflows.reduce((acc, workflow) => {
    const module = workflow.module
    if (!acc[module]) {
      acc[module] = []
    }
    acc[module].push(workflow)
    return acc
  }, {} as Record<string, WorkflowWithStats[]>)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Workflows</h1>
          <p className="text-sm text-gray-500">
            Configure fluxos de trabalho automatizados para cada módulo
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Link
            href="/dashboard/settings/workflows/templates"
            className="inline-flex items-center gap-2 px-4 py-2 text-gray-600 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            <Mail className="h-4 w-4" />
            Modelos
          </Link>
          <Link
            href="/dashboard/settings/workflows/logs"
            className="inline-flex items-center gap-2 px-4 py-2 text-gray-600 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            <Activity className="h-4 w-4" />
            Ver Logs
          </Link>
          {enabledModules.includes('construction') && (
            <button
              onClick={handleInitDefaults}
              disabled={initDefaultWorkflows.isPending}
              className="inline-flex items-center gap-2 px-4 py-2 text-primary-600 bg-primary-50 border border-primary-200 rounded-md hover:bg-primary-100 disabled:opacity-50"
            >
              {initDefaultWorkflows.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Wand2 className="h-4 w-4" />
              )}
              Criar Padrões
            </button>
          )}
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
          >
            <Plus className="h-4 w-4" />
            Novo Workflow
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
        </div>
      ) : workflows.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <Settings className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900">Nenhum workflow</h3>
          <p className="mt-1 text-sm text-gray-500">
            Crie um workflow para automatizar processos
          </p>
          <div className="mt-4 flex items-center justify-center gap-3">
            {enabledModules.includes('construction') && (
              <button
                onClick={handleInitDefaults}
                disabled={initDefaultWorkflows.isPending}
                className="inline-flex items-center gap-2 px-4 py-2 text-primary-600 bg-primary-50 border border-primary-200 rounded-md hover:bg-primary-100 disabled:opacity-50"
              >
                {initDefaultWorkflows.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Wand2 className="h-4 w-4" />
                )}
                Criar Padrões Construção
              </button>
            )}
            <button
              onClick={() => setIsCreateModalOpen(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
            >
              <Plus className="h-4 w-4" />
              Criar Workflow Manual
            </button>
          </div>
        </div>
      ) : (
        <div className="space-y-8">
          {Object.entries(groupedWorkflows).map(([module, moduleWorkflows]) => (
            <div key={module}>
              <h2 className="text-lg font-medium text-gray-900 mb-4">
                {moduleLabels[module] || module}
              </h2>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {moduleWorkflows.map((workflow) => (
                  <WorkflowCard
                    key={workflow.id}
                    workflow={workflow}
                    onEdit={() => handleEdit(workflow)}
                    onDuplicate={() => setDuplicatingWorkflow(workflow)}
                    onDelete={() => handleDelete(workflow.id)}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Workflow Modal */}
      {isCreateModalOpen && (
        <CreateWorkflowModal
          enabledModules={enabledModules}
          onClose={() => setIsCreateModalOpen(false)}
        />
      )}

      {/* Duplicate Workflow Modal */}
      {duplicatingWorkflow && (
        <DuplicateWorkflowModal
          workflow={duplicatingWorkflow}
          onClose={() => setDuplicatingWorkflow(null)}
        />
      )}
    </div>
  )
}

function WorkflowCard({
  workflow,
  onEdit,
  onDuplicate,
  onDelete,
}: {
  workflow: WorkflowWithStats
  onEdit: () => void
  onDuplicate: () => void
  onDelete: () => void
}) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="text-base font-medium text-gray-900 truncate">
              {workflow.name}
            </h3>
            {workflow.is_default && (
              <span className="px-2 py-0.5 text-xs bg-primary-100 text-primary-700 rounded-full">
                Padrão
              </span>
            )}
          </div>
          <p className="text-sm text-gray-500 mt-1">
            {entityTypeLabels[workflow.entity_type]}
          </p>
        </div>
        <div className="flex items-center gap-1">
          {workflow.is_active ? (
            <CheckCircle className="h-5 w-5 text-green-500" />
          ) : (
            <XCircle className="h-5 w-5 text-gray-400" />
          )}
        </div>
      </div>

      {workflow.description && (
        <p className="mt-2 text-sm text-gray-600 line-clamp-2">
          {workflow.description}
        </p>
      )}

      <div className="mt-3 flex items-center gap-4 text-xs text-gray-500">
        <span>{workflow.state_count} estados</span>
        <span>{workflow.trigger_count} gatilhos</span>
        <span>{workflow.action_count} ações</span>
      </div>

      <div className="mt-4 flex items-center gap-2 pt-3 border-t border-gray-100">
        <button
          onClick={onEdit}
          className="flex-1 px-3 py-1.5 text-sm text-primary-600 hover:bg-primary-50 rounded-md"
        >
          Editar
        </button>
        <button
          onClick={onDuplicate}
          className="p-1.5 text-gray-400 hover:text-gray-600 rounded-md"
          title="Duplicar"
        >
          <Copy className="h-4 w-4" />
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

function CreateWorkflowModal({
  enabledModules,
  onClose,
}: {
  enabledModules: string[]
  onClose: () => void
}) {
  const router = useRouter()
  const createWorkflow = useCreateWorkflow()

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<WorkflowFormData>({
    resolver: zodResolver(workflowSchema),
    defaultValues: {
      module: enabledModules.includes('appointments') ? 'appointments' : 'construction',
      entity_type: 'session',
      is_default: false,
    },
  })

  const module = watch('module')

  const entityTypeOptions = module === 'appointments'
    ? [{ value: 'session', label: 'Sessão' }]
    : [
        { value: 'budget', label: 'Orçamento' },
        { value: 'project', label: 'Projeto' },
      ]

  const onSubmit = async (data: WorkflowFormData) => {
    try {
      const workflow = await createWorkflow.mutateAsync(data)
      onClose()
      router.push(`/dashboard/settings/workflows/${workflow.id}`)
    } catch (error) {
      console.error('Failed to create workflow:', error)
    }
  }

  return (
    <Modal isOpen={true} onClose={onClose} title="Novo Workflow">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Nome *
          </label>
          <input
            type="text"
            {...register('name')}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            placeholder="Ex: Fluxo de Sessão"
          />
          {errors.name && (
            <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Descrição
          </label>
          <textarea
            {...register('description')}
            rows={2}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            placeholder="Descrição do workflow"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Módulo *
            </label>
            <select
              {...register('module')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              {enabledModules.includes('appointments') && (
                <option value="appointments">Consultas</option>
              )}
              {enabledModules.includes('construction') && (
                <option value="construction">Construção</option>
              )}
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Tipo de Entidade *
            </label>
            <select
              {...register('entity_type')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            >
              {entityTypeOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="flex items-center">
          <input
            type="checkbox"
            {...register('is_default')}
            id="is_default"
            className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
          />
          <label htmlFor="is_default" className="ml-2 text-sm text-gray-700">
            Definir como workflow padrão
          </label>
        </div>

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
            Criar Workflow
          </button>
        </div>
      </form>
    </Modal>
  )
}

function DuplicateWorkflowModal({
  workflow,
  onClose,
}: {
  workflow: WorkflowWithStats
  onClose: () => void
}) {
  const router = useRouter()
  const duplicateWorkflow = useDuplicateWorkflow()
  const [name, setName] = useState(`${workflow.name} (cópia)`)

  const handleDuplicate = async () => {
    try {
      const newWorkflow = await duplicateWorkflow.mutateAsync({ id: workflow.id, name })
      onClose()
      router.push(`/dashboard/settings/workflows/${newWorkflow.id}`)
    } catch (error) {
      console.error('Failed to duplicate workflow:', error)
    }
  }

  return (
    <Modal isOpen={true} onClose={onClose} title="Duplicar Workflow">
      <div className="space-y-4">
        <p className="text-sm text-gray-600">
          Uma cópia de "{workflow.name}" será criada com todos os estados, gatilhos e ações.
        </p>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Nome do Novo Workflow
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
          />
        </div>

        <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Cancelar
          </button>
          <button
            onClick={handleDuplicate}
            disabled={!name.trim() || duplicateWorkflow.isPending}
            className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 disabled:opacity-50"
          >
            {duplicateWorkflow.isPending && (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            )}
            Duplicar
          </button>
        </div>
      </div>
    </Modal>
  )
}
