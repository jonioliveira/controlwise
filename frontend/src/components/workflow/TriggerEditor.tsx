'use client'

import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Modal } from '@/components/ui/Modal'
import { triggerSchema, triggerTypeLabels, actionTypeLabels } from '@/schemas/workflow'
import {
  useCreateTrigger,
  useUpdateTrigger,
  useCreateAction,
  useDeleteAction,
  useTemplates,
} from '@/hooks/useWorkflows'
import { Loader2, Plus, Trash2 } from 'lucide-react'
import type { WorkflowTrigger, WorkflowAction, WorkflowEntityType } from '@/types'
import { z } from 'zod'

interface TriggerEditorProps {
  workflowId: string
  stateId: string
  trigger: WorkflowTrigger | null
  entityType: WorkflowEntityType
  onClose: () => void
}

type TriggerFormData = z.infer<typeof triggerSchema>

export function TriggerEditor({ workflowId, stateId, trigger, entityType, onClose }: TriggerEditorProps) {
  const isEditing = !!trigger
  const [isAddingAction, setIsAddingAction] = useState(false)

  const createTrigger = useCreateTrigger()
  const updateTrigger = useUpdateTrigger()
  const createAction = useCreateAction()
  const deleteAction = useDeleteAction()
  const { data: templates = [] } = useTemplates()

  const {
    register,
    handleSubmit,
    watch,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<TriggerFormData>({
    resolver: zodResolver(triggerSchema),
    defaultValues: {
      state_id: stateId,
      trigger_type: 'on_enter',
      time_offset_minutes: undefined,
      time_field: 'scheduled_at',
    },
  })

  const triggerType = watch('trigger_type')
  const isTimeBased = triggerType === 'time_before' || triggerType === 'time_after'
  const isRecurring = triggerType === 'recurring'

  useEffect(() => {
    if (trigger) {
      reset({
        state_id: trigger.state_id || stateId,
        transition_id: trigger.transition_id || undefined,
        trigger_type: trigger.trigger_type,
        time_offset_minutes: trigger.time_offset_minutes || undefined,
        time_field: trigger.time_field || 'scheduled_at',
        recurring_cron: trigger.recurring_cron || undefined,
      })
    }
  }, [trigger, stateId, reset])

  const onSubmit = async (data: TriggerFormData) => {
    try {
      // Clean up data based on trigger type
      const cleanData = {
        ...data,
        state_id: stateId,
        time_offset_minutes: isTimeBased ? data.time_offset_minutes : undefined,
        time_field: isTimeBased ? data.time_field : undefined,
        recurring_cron: isRecurring ? data.recurring_cron : undefined,
      }

      if (isEditing && trigger) {
        await updateTrigger.mutateAsync({
          triggerId: trigger.id,
          workflowId,
          data: cleanData,
        })
      } else {
        await createTrigger.mutateAsync({
          workflowId,
          data: cleanData,
        })
      }
      onClose()
    } catch (error) {
      console.error('Failed to save trigger:', error)
    }
  }

  const handleAddAction = async (actionType: string, templateId?: string) => {
    if (!trigger) return

    try {
      await createAction.mutateAsync({
        triggerId: trigger.id,
        workflowId,
        data: {
          action_type: actionType as 'send_whatsapp' | 'send_email' | 'update_field' | 'create_task',
          action_order: (trigger.actions?.length || 0),
          template_id: templateId,
        },
      })
      setIsAddingAction(false)
    } catch (error) {
      console.error('Failed to add action:', error)
    }
  }

  const handleDeleteAction = async (actionId: string) => {
    if (confirm('Tem certeza que deseja eliminar esta ação?')) {
      await deleteAction.mutateAsync({ actionId, workflowId })
    }
  }

  return (
    <Modal
      isOpen={true}
      onClose={onClose}
      title={isEditing ? 'Editar Gatilho' : 'Novo Gatilho'}
      size="lg"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Tipo de Gatilho *
          </label>
          <select
            {...register('trigger_type')}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
          >
            {Object.entries(triggerTypeLabels).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
          {errors.trigger_type && (
            <p className="mt-1 text-sm text-red-600">{errors.trigger_type.message}</p>
          )}
        </div>

        {isTimeBased && (
          <>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Tempo (minutos) *
                </label>
                <input
                  type="number"
                  {...register('time_offset_minutes', { valueAsNumber: true })}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  placeholder="Ex: 1440 (24 horas)"
                />
                <p className="mt-1 text-xs text-gray-500">
                  {triggerType === 'time_before' ? 'Minutos antes do evento' : 'Minutos depois do evento'}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Campo de Referência
                </label>
                <select
                  {...register('time_field')}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                >
                  {entityType === 'session' && (
                    <option value="scheduled_at">Data/hora agendada</option>
                  )}
                  <option value="created_at">Data de criação</option>
                </select>
              </div>
            </div>

            <div className="p-3 bg-blue-50 rounded-md">
              <p className="text-sm text-blue-700">
                Exemplo: Para enviar um lembrete 24h antes, selecione "Tempo antes" e coloque 1440 minutos.
              </p>
            </div>
          </>
        )}

        {isRecurring && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Expressão Cron *
            </label>
            <input
              type="text"
              {...register('recurring_cron')}
              className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="0 10 * * 0 (Domingos às 10:00)"
            />
            <p className="mt-1 text-xs text-gray-500">
              Formato: minuto hora dia-mês mês dia-semana
            </p>
          </div>
        )}

        {/* Actions section - only show when editing */}
        {isEditing && trigger && (
          <div className="pt-4 border-t border-gray-200">
            <div className="flex items-center justify-between mb-3">
              <h4 className="text-sm font-medium text-gray-900">Ações</h4>
              <button
                type="button"
                onClick={() => setIsAddingAction(true)}
                className="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700"
              >
                <Plus className="h-4 w-4" />
                Adicionar Ação
              </button>
            </div>

            {trigger.actions && trigger.actions.length > 0 ? (
              <div className="space-y-2">
                {trigger.actions.map((action) => (
                  <ActionItem
                    key={action.id}
                    action={action}
                    templates={templates}
                    onDelete={() => handleDeleteAction(action.id)}
                  />
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500 text-center py-3">
                Nenhuma ação configurada
              </p>
            )}

            {/* Add action form */}
            {isAddingAction && (
              <AddActionForm
                templates={templates}
                onAdd={handleAddAction}
                onCancel={() => setIsAddingAction(false)}
              />
            )}
          </div>
        )}

        <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            {isEditing ? 'Fechar' : 'Cancelar'}
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 disabled:opacity-50"
          >
            {isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
            {isEditing ? 'Guardar' : 'Criar Gatilho'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

function ActionItem({
  action,
  templates,
  onDelete,
}: {
  action: WorkflowAction
  templates: { id: string; name: string; channel: string }[]
  onDelete: () => void
}) {
  const template = templates.find((t) => t.id === action.template_id)

  return (
    <div className="flex items-center justify-between p-3 bg-gray-50 rounded-md">
      <div>
        <p className="text-sm font-medium text-gray-900">
          {actionTypeLabels[action.action_type]}
        </p>
        {template && (
          <p className="text-xs text-gray-500">Modelo: {template.name}</p>
        )}
      </div>
      <button
        type="button"
        onClick={onDelete}
        className="p-1 text-gray-400 hover:text-red-600"
      >
        <Trash2 className="h-4 w-4" />
      </button>
    </div>
  )
}

function AddActionForm({
  templates,
  onAdd,
  onCancel,
}: {
  templates: { id: string; name: string; channel: string }[]
  onAdd: (actionType: string, templateId?: string) => void
  onCancel: () => void
}) {
  const [actionType, setActionType] = useState('')
  const [templateId, setTemplateId] = useState('')

  const needsTemplate = actionType === 'send_whatsapp' || actionType === 'send_email'
  const filteredTemplates = templates.filter(
    (t) => t.channel === (actionType === 'send_whatsapp' ? 'whatsapp' : 'email')
  )

  return (
    <div className="mt-3 p-3 bg-gray-100 rounded-md space-y-3">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Tipo de Ação
        </label>
        <select
          value={actionType}
          onChange={(e) => {
            setActionType(e.target.value)
            setTemplateId('')
          }}
          className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
        >
          <option value="">Selecione...</option>
          {Object.entries(actionTypeLabels).map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </select>
      </div>

      {needsTemplate && (
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Modelo de Mensagem
          </label>
          <select
            value={templateId}
            onChange={(e) => setTemplateId(e.target.value)}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
          >
            <option value="">Selecione um modelo...</option>
            {filteredTemplates.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
          {filteredTemplates.length === 0 && (
            <p className="mt-1 text-xs text-amber-600">
              Nenhum modelo disponível. Crie um modelo primeiro.
            </p>
          )}
        </div>
      )}

      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800"
        >
          Cancelar
        </button>
        <button
          type="button"
          onClick={() => onAdd(actionType, templateId || undefined)}
          disabled={!actionType || (needsTemplate && !templateId)}
          className="px-3 py-1.5 text-sm bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50"
        >
          Adicionar
        </button>
      </div>
    </div>
  )
}
