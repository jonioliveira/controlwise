'use client'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Modal } from '@/components/ui/Modal'
import { stateSchema, type StateFormData, stateTypeLabels, defaultStateColors } from '@/schemas/workflow'
import { useCreateState, useUpdateState } from '@/hooks/useWorkflows'
import { Loader2 } from 'lucide-react'
import type { WorkflowState } from '@/types'

interface StateEditorProps {
  workflowId: string
  state: WorkflowState | null
  nextPosition: number
  onClose: () => void
}

export function StateEditor({ workflowId, state, nextPosition, onClose }: StateEditorProps) {
  const isEditing = !!state
  const createState = useCreateState()
  const updateState = useUpdateState()

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<StateFormData>({
    resolver: zodResolver(stateSchema),
    defaultValues: {
      name: '',
      display_name: '',
      description: '',
      state_type: 'intermediate',
      color: defaultStateColors.intermediate,
      position: nextPosition,
    },
  })

  const stateType = watch('state_type')

  useEffect(() => {
    if (state) {
      reset({
        name: state.name,
        display_name: state.display_name,
        description: state.description || '',
        state_type: state.state_type,
        color: state.color || defaultStateColors[state.state_type],
        position: state.position,
      })
    }
  }, [state, reset])

  useEffect(() => {
    // Update color when state type changes (only if not editing)
    if (!isEditing && stateType) {
      setValue('color', defaultStateColors[stateType])
    }
  }, [stateType, isEditing, setValue])

  const onSubmit = async (data: StateFormData) => {
    try {
      if (isEditing && state) {
        await updateState.mutateAsync({
          workflowId,
          stateId: state.id,
          data,
        })
      } else {
        await createState.mutateAsync({
          workflowId,
          data: {
            ...data,
            position: nextPosition,
          },
        })
      }
      onClose()
    } catch (error) {
      console.error('Failed to save state:', error)
    }
  }

  // Auto-generate name from display_name
  const handleDisplayNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const displayName = e.target.value
    if (!isEditing) {
      const name = displayName
        .toLowerCase()
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .replace(/[^a-z0-9\s]/g, '')
        .replace(/\s+/g, '_')
        .substring(0, 50)
      setValue('name', name)
    }
  }

  return (
    <Modal
      isOpen={true}
      onClose={onClose}
      title={isEditing ? 'Editar Estado' : 'Novo Estado'}
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Nome de Exibição *
          </label>
          <input
            type="text"
            {...register('display_name')}
            onChange={(e) => {
              register('display_name').onChange(e)
              handleDisplayNameChange(e)
            }}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
            placeholder="Ex: Agendado, Em Revisão, Concluído"
          />
          {errors.display_name && (
            <p className="mt-1 text-sm text-red-600">{errors.display_name.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Identificador *
          </label>
          <input
            type="text"
            {...register('name')}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 bg-gray-50"
            placeholder="ex: agendado"
            readOnly={isEditing}
          />
          <p className="mt-1 text-xs text-gray-500">
            Identificador único (apenas letras minúsculas e underscores)
          </p>
          {errors.name && (
            <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Tipo de Estado *
          </label>
          <select
            {...register('state_type')}
            className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
          >
            {Object.entries(stateTypeLabels).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
          <p className="mt-1 text-xs text-gray-500">
            Inicial: primeiro estado do fluxo. Final: estado de conclusão.
          </p>
          {errors.state_type && (
            <p className="mt-1 text-sm text-red-600">{errors.state_type.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Cor
          </label>
          <div className="flex items-center gap-2">
            <input
              type="color"
              {...register('color')}
              className="h-10 w-14 rounded border border-gray-300 cursor-pointer"
            />
            <input
              type="text"
              {...register('color')}
              className="flex-1 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              placeholder="#3B82F6"
            />
          </div>
          {errors.color && (
            <p className="mt-1 text-sm text-red-600">{errors.color.message}</p>
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
            placeholder="Descrição opcional do estado"
          />
          {errors.description && (
            <p className="mt-1 text-sm text-red-600">{errors.description.message}</p>
          )}
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
            {isEditing ? 'Guardar' : 'Criar Estado'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
