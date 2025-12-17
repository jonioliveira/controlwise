'use client'

import { use } from 'react'
import { useRouter } from 'next/navigation'
import { ArrowLeft, Loader2, AlertCircle, ToggleLeft, ToggleRight, Trash2 } from 'lucide-react'
import { useWorkflow, useUpdateWorkflow, useDeleteWorkflow } from '@/hooks/useWorkflows'
import { WorkflowBuilder } from '@/components/workflow'
import { useState } from 'react'
import { Modal } from '@/components/ui/Modal'

const moduleLabels: Record<string, string> = {
  appointments: 'Consultas',
  construction: 'Construção',
}

const entityTypeLabels: Record<string, string> = {
  session: 'Sessão',
  budget: 'Orçamento',
  project: 'Projeto',
}

interface WorkflowEditorPageProps {
  params: Promise<{ id: string }>
}

export default function WorkflowEditorPage({ params }: WorkflowEditorPageProps) {
  const { id } = use(params)
  const router = useRouter()
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)

  const { data: workflow, isLoading, error } = useWorkflow(id)
  const updateWorkflow = useUpdateWorkflow()
  const deleteWorkflow = useDeleteWorkflow()

  const handleBack = () => {
    router.push('/dashboard/settings/workflows')
  }

  const handleToggleActive = async () => {
    if (!workflow) return
    await updateWorkflow.mutateAsync({
      id: workflow.id,
      data: { is_active: !workflow.is_active },
    })
  }

  const handleToggleDefault = async () => {
    if (!workflow) return
    await updateWorkflow.mutateAsync({
      id: workflow.id,
      data: { is_default: !workflow.is_default },
    })
  }

  const handleDelete = async () => {
    if (!workflow) return
    try {
      await deleteWorkflow.mutateAsync(workflow.id)
      router.push('/dashboard/settings/workflows')
    } catch (error) {
      console.error('Failed to delete workflow:', error)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
      </div>
    )
  }

  if (error || !workflow) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] text-center">
        <AlertCircle className="h-12 w-12 text-red-500 mb-4" />
        <h2 className="text-lg font-medium text-gray-900">Erro ao carregar workflow</h2>
        <p className="mt-1 text-sm text-gray-500">
          {error?.message || 'Workflow não encontrado'}
        </p>
        <button
          onClick={handleBack}
          className="mt-4 inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-primary-600 hover:text-primary-700"
        >
          <ArrowLeft className="h-4 w-4" />
          Voltar para Workflows
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={handleBack}
            className="p-2 text-gray-400 hover:text-gray-600 rounded-md hover:bg-gray-100"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-gray-900">{workflow.name}</h1>
              {workflow.is_default && (
                <span className="px-2 py-0.5 text-xs bg-primary-100 text-primary-700 rounded-full">
                  Padrão
                </span>
              )}
              {!workflow.is_active && (
                <span className="px-2 py-0.5 text-xs bg-gray-100 text-gray-600 rounded-full">
                  Inativo
                </span>
              )}
            </div>
            <p className="text-sm text-gray-500">
              {moduleLabels[workflow.module]} • {entityTypeLabels[workflow.entity_type]}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          {/* Toggle Active */}
          <button
            onClick={handleToggleActive}
            disabled={updateWorkflow.isPending}
            className={`inline-flex items-center gap-2 px-3 py-1.5 text-sm rounded-md border transition-colors ${
              workflow.is_active
                ? 'border-green-200 bg-green-50 text-green-700 hover:bg-green-100'
                : 'border-gray-200 bg-gray-50 text-gray-600 hover:bg-gray-100'
            }`}
          >
            {workflow.is_active ? (
              <>
                <ToggleRight className="h-4 w-4" />
                Ativo
              </>
            ) : (
              <>
                <ToggleLeft className="h-4 w-4" />
                Inativo
              </>
            )}
          </button>

          {/* Toggle Default */}
          <button
            onClick={handleToggleDefault}
            disabled={updateWorkflow.isPending || workflow.is_default}
            className={`px-3 py-1.5 text-sm rounded-md border transition-colors ${
              workflow.is_default
                ? 'border-primary-200 bg-primary-50 text-primary-700'
                : 'border-gray-200 bg-white text-gray-600 hover:bg-gray-50'
            } disabled:opacity-50`}
            title={workflow.is_default ? 'Este é o workflow padrão' : 'Definir como padrão'}
          >
            {workflow.is_default ? 'Padrão' : 'Definir Padrão'}
          </button>

          {/* Delete */}
          <button
            onClick={() => setIsDeleteModalOpen(true)}
            className="p-2 text-gray-400 hover:text-red-600 rounded-md hover:bg-red-50"
            title="Eliminar workflow"
          >
            <Trash2 className="h-5 w-5" />
          </button>
        </div>
      </div>

      {/* Workflow Builder */}
      <WorkflowBuilder workflow={workflow} />

      {/* Delete Confirmation Modal */}
      {isDeleteModalOpen && (
        <Modal
          isOpen={true}
          onClose={() => setIsDeleteModalOpen(false)}
          title="Eliminar Workflow"
        >
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              Tem certeza que deseja eliminar o workflow "{workflow.name}"?
            </p>
            <p className="text-sm text-red-600">
              Esta ação não pode ser desfeita. Todos os estados, gatilhos e ações serão eliminados.
            </p>

            <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
              <button
                type="button"
                onClick={() => setIsDeleteModalOpen(false)}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
              >
                Cancelar
              </button>
              <button
                onClick={handleDelete}
                disabled={deleteWorkflow.isPending}
                className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-red-600 border border-transparent rounded-md hover:bg-red-700 disabled:opacity-50"
              >
                {deleteWorkflow.isPending && (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                )}
                Eliminar
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
