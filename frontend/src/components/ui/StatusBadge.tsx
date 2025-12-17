'use client'

import type {
  WorkSheetStatus,
  BudgetStatus,
  ProjectStatus,
  TaskStatus,
  PaymentStatus,
} from '@/types'

type Status = WorkSheetStatus | BudgetStatus | ProjectStatus | TaskStatus | PaymentStatus

const statusConfig: Record<
  Status,
  { label: string; className: string }
> = {
  // WorkSheet
  draft: { label: 'Rascunho', className: 'bg-gray-100 text-gray-800' },
  under_review: { label: 'Em Revisão', className: 'bg-yellow-100 text-yellow-800' },
  approved: { label: 'Aprovado', className: 'bg-green-100 text-green-800' },

  // Budget
  sent: { label: 'Enviado', className: 'bg-blue-100 text-blue-800' },
  rejected: { label: 'Rejeitado', className: 'bg-red-100 text-red-800' },
  expired: { label: 'Expirado', className: 'bg-gray-100 text-gray-800' },

  // Project
  in_progress: { label: 'Em Progresso', className: 'bg-blue-100 text-blue-800' },
  on_hold: { label: 'Em Espera', className: 'bg-yellow-100 text-yellow-800' },
  completed: { label: 'Concluído', className: 'bg-green-100 text-green-800' },
  cancelled: { label: 'Cancelado', className: 'bg-red-100 text-red-800' },

  // Task
  todo: { label: 'Por Fazer', className: 'bg-gray-100 text-gray-800' },

  // Payment
  pending: { label: 'Pendente', className: 'bg-yellow-100 text-yellow-800' },
  paid: { label: 'Pago', className: 'bg-green-100 text-green-800' },
  overdue: { label: 'Atrasado', className: 'bg-red-100 text-red-800' },
}

interface StatusBadgeProps {
  status: Status
  className?: string
}

export function StatusBadge({ status, className = '' }: StatusBadgeProps) {
  const config = statusConfig[status] || {
    label: status,
    className: 'bg-gray-100 text-gray-800',
  }

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.className} ${className}`}
    >
      {config.label}
    </span>
  )
}

// Priority badge for tasks
type Priority = 'low' | 'medium' | 'high' | 'urgent'

const priorityConfig: Record<Priority, { label: string; className: string }> = {
  low: { label: 'Baixa', className: 'bg-gray-100 text-gray-800' },
  medium: { label: 'Média', className: 'bg-blue-100 text-blue-800' },
  high: { label: 'Alta', className: 'bg-orange-100 text-orange-800' },
  urgent: { label: 'Urgente', className: 'bg-red-100 text-red-800' },
}

interface PriorityBadgeProps {
  priority: Priority
  className?: string
}

export function PriorityBadge({ priority, className = '' }: PriorityBadgeProps) {
  const config = priorityConfig[priority]

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.className} ${className}`}
    >
      {config.label}
    </span>
  )
}
