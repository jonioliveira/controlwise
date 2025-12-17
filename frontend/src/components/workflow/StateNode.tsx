'use client'

import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Settings, GripVertical, Trash2, ChevronDown, ChevronUp } from 'lucide-react'
import type { WorkflowState, WorkflowTrigger } from '@/types'
import { stateTypeLabels, triggerTypeLabels } from '@/schemas/workflow'
import { useState } from 'react'

interface StateNodeProps {
  state: WorkflowState
  isFirst: boolean
  isLast: boolean
  onEdit: (state: WorkflowState) => void
  onDelete: (stateId: string) => void
  onAddTrigger: (stateId: string) => void
  onEditTrigger: (trigger: WorkflowTrigger) => void
  onDeleteTrigger: (triggerId: string) => void
}

export function StateNode({
  state,
  isFirst,
  isLast,
  onEdit,
  onDelete,
  onAddTrigger,
  onEditTrigger,
  onDeleteTrigger,
}: StateNodeProps) {
  const [expanded, setExpanded] = useState(false)
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: state.id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const stateColor = state.color || '#3B82F6'
  const triggers = state.triggers || []

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`relative ${!isLast ? 'pb-4' : ''}`}
    >
      {/* Connection line to next state */}
      {!isLast && (
        <div className="absolute left-1/2 -translate-x-1/2 top-full -mt-4 w-0.5 h-8 bg-gray-300" />
      )}

      <div
        className="bg-white rounded-lg border-2 shadow-sm hover:shadow-md transition-shadow"
        style={{ borderColor: stateColor }}
      >
        {/* Header */}
        <div className="flex items-center gap-2 p-3 border-b border-gray-100">
          <button
            {...attributes}
            {...listeners}
            className="cursor-grab active:cursor-grabbing text-gray-400 hover:text-gray-600"
          >
            <GripVertical className="h-4 w-4" />
          </button>

          <div
            className="w-3 h-3 rounded-full"
            style={{ backgroundColor: stateColor }}
          />

          <div className="flex-1 min-w-0">
            <h4 className="text-sm font-medium text-gray-900 truncate">
              {state.display_name}
            </h4>
            <p className="text-xs text-gray-500">
              {stateTypeLabels[state.state_type]}
            </p>
          </div>

          <div className="flex items-center gap-1">
            <button
              onClick={() => onEdit(state)}
              className="p-1 text-gray-400 hover:text-gray-600 rounded"
              title="Editar estado"
            >
              <Settings className="h-4 w-4" />
            </button>
            {!isFirst && (
              <button
                onClick={() => onDelete(state.id)}
                className="p-1 text-gray-400 hover:text-red-600 rounded"
                title="Eliminar estado"
              >
                <Trash2 className="h-4 w-4" />
              </button>
            )}
          </div>
        </div>

        {/* Triggers section */}
        {triggers.length > 0 && (
          <div className="px-3 py-2">
            <button
              onClick={() => setExpanded(!expanded)}
              className="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-700"
            >
              {expanded ? (
                <ChevronUp className="h-3 w-3" />
              ) : (
                <ChevronDown className="h-3 w-3" />
              )}
              {triggers.length} gatilho{triggers.length !== 1 ? 's' : ''}
            </button>

            {expanded && (
              <div className="mt-2 space-y-1">
                {triggers.map((trigger) => (
                  <div
                    key={trigger.id}
                    className="flex items-center justify-between p-2 bg-gray-50 rounded text-xs"
                  >
                    <span className="text-gray-700">
                      {triggerTypeLabels[trigger.trigger_type]}
                      {trigger.time_offset_minutes && (
                        <span className="text-gray-500">
                          {' '}({Math.abs(trigger.time_offset_minutes)} min)
                        </span>
                      )}
                    </span>
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => onEditTrigger(trigger)}
                        className="p-0.5 text-gray-400 hover:text-gray-600"
                      >
                        <Settings className="h-3 w-3" />
                      </button>
                      <button
                        onClick={() => onDeleteTrigger(trigger.id)}
                        className="p-0.5 text-gray-400 hover:text-red-600"
                      >
                        <Trash2 className="h-3 w-3" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Add trigger button */}
        <div className="px-3 pb-3">
          <button
            onClick={() => onAddTrigger(state.id)}
            className="w-full py-1.5 text-xs text-primary-600 hover:text-primary-700 hover:bg-primary-50 rounded border border-dashed border-primary-300"
          >
            + Adicionar Gatilho
          </button>
        </div>
      </div>

      {/* Arrow to next state */}
      {!isLast && (
        <div className="absolute left-1/2 -translate-x-1/2 top-full mt-2">
          <div className="w-0 h-0 border-l-4 border-r-4 border-t-4 border-l-transparent border-r-transparent border-t-gray-300" />
        </div>
      )}
    </div>
  )
}
