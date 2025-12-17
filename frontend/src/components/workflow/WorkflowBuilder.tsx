'use client'

import { useState, useCallback } from 'react'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { Plus, Save } from 'lucide-react'
import { StateNode } from './StateNode'
import { StateEditor } from './StateEditor'
import { TriggerEditor } from './TriggerEditor'
import {
  useReorderStates,
  useDeleteState,
  useDeleteTrigger,
} from '@/hooks/useWorkflows'
import type { Workflow, WorkflowState, WorkflowTrigger } from '@/types'

interface WorkflowBuilderProps {
  workflow: Workflow
  onSave?: () => void
}

export function WorkflowBuilder({ workflow, onSave }: WorkflowBuilderProps) {
  const [editingState, setEditingState] = useState<WorkflowState | null>(null)
  const [isAddingState, setIsAddingState] = useState(false)
  const [editingTrigger, setEditingTrigger] = useState<{
    trigger: WorkflowTrigger | null
    stateId: string
  } | null>(null)

  const reorderStates = useReorderStates()
  const deleteState = useDeleteState()
  const deleteTrigger = useDeleteTrigger()

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const states = [...(workflow.states || [])].sort((a, b) => a.position - b.position)

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event

      if (over && active.id !== over.id) {
        const oldIndex = states.findIndex((s) => s.id === active.id)
        const newIndex = states.findIndex((s) => s.id === over.id)

        // Reorder locally
        const newStates = [...states]
        const [removed] = newStates.splice(oldIndex, 1)
        newStates.splice(newIndex, 0, removed)

        // Update positions
        const stateIds = newStates.map((s) => s.id)
        reorderStates.mutate({ workflowId: workflow.id, stateIds })
      }
    },
    [states, workflow.id, reorderStates]
  )

  const handleEditState = (state: WorkflowState) => {
    setEditingState(state)
    setIsAddingState(false)
  }

  const handleAddState = () => {
    setEditingState(null)
    setIsAddingState(true)
  }

  const handleDeleteState = (stateId: string) => {
    if (confirm('Tem certeza que deseja eliminar este estado?')) {
      deleteState.mutate({ workflowId: workflow.id, stateId })
    }
  }

  const handleAddTrigger = (stateId: string) => {
    setEditingTrigger({ trigger: null, stateId })
  }

  const handleEditTrigger = (trigger: WorkflowTrigger) => {
    setEditingTrigger({ trigger, stateId: trigger.state_id || '' })
  }

  const handleDeleteTrigger = (triggerId: string) => {
    if (confirm('Tem certeza que deseja eliminar este gatilho?')) {
      deleteTrigger.mutate({ triggerId, workflowId: workflow.id })
    }
  }

  const handleCloseStateEditor = () => {
    setEditingState(null)
    setIsAddingState(false)
  }

  const handleCloseTriggerEditor = () => {
    setEditingTrigger(null)
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold text-gray-900">{workflow.name}</h2>
          <p className="text-sm text-gray-500">
            {workflow.description || 'Sem descrição'}
          </p>
        </div>
        {onSave && (
          <button
            onClick={onSave}
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
          >
            <Save className="h-4 w-4" />
            Guardar
          </button>
        )}
      </div>

      {/* States flow */}
      <div className="bg-gray-50 rounded-lg p-6">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={states.map((s) => s.id)}
            strategy={verticalListSortingStrategy}
          >
            <div className="max-w-md mx-auto space-y-2">
              {states.map((state, index) => (
                <StateNode
                  key={state.id}
                  state={state}
                  isFirst={index === 0}
                  isLast={index === states.length - 1}
                  onEdit={handleEditState}
                  onDelete={handleDeleteState}
                  onAddTrigger={handleAddTrigger}
                  onEditTrigger={handleEditTrigger}
                  onDeleteTrigger={handleDeleteTrigger}
                />
              ))}
            </div>
          </SortableContext>
        </DndContext>

        {/* Add state button */}
        <div className="max-w-md mx-auto mt-6">
          <button
            onClick={handleAddState}
            className="w-full py-3 flex items-center justify-center gap-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-white rounded-lg border-2 border-dashed border-gray-300 hover:border-gray-400 transition-colors"
          >
            <Plus className="h-4 w-4" />
            Adicionar Estado
          </button>
        </div>
      </div>

      {/* State Editor Modal */}
      {(editingState || isAddingState) && (
        <StateEditor
          workflowId={workflow.id}
          state={editingState}
          nextPosition={states.length}
          onClose={handleCloseStateEditor}
        />
      )}

      {/* Trigger Editor Modal */}
      {editingTrigger && (
        <TriggerEditor
          workflowId={workflow.id}
          stateId={editingTrigger.stateId}
          trigger={editingTrigger.trigger}
          entityType={workflow.entity_type}
          onClose={handleCloseTriggerEditor}
        />
      )}
    </div>
  )
}
