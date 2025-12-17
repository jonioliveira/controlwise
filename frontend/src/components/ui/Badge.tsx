'use client'

import { cn } from '@/lib/utils'

type BadgeVariant = 'default' | 'success' | 'warning' | 'error' | 'info'

interface BadgeProps {
  children: React.ReactNode
  variant?: BadgeVariant
  className?: string
}

const variantStyles: Record<BadgeVariant, string> = {
  default: 'bg-gray-100 text-gray-800',
  success: 'bg-green-100 text-green-800',
  warning: 'bg-yellow-100 text-yellow-800',
  error: 'bg-red-100 text-red-800',
  info: 'bg-blue-100 text-blue-800',
}

export function Badge({ children, variant = 'default', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
        variantStyles[variant],
        className
      )}
    >
      {children}
    </span>
  )
}

// Status-specific badges for common use cases
export function StatusBadge({
  status,
  labels
}: {
  status: string
  labels?: Record<string, { label: string; variant: BadgeVariant }>
}) {
  const defaultLabels: Record<string, { label: string; variant: BadgeVariant }> = {
    // Session statuses
    pending: { label: 'Pendente', variant: 'warning' },
    confirmed: { label: 'Confirmada', variant: 'info' },
    cancelled: { label: 'Cancelada', variant: 'error' },
    completed: { label: 'Concluída', variant: 'success' },
    no_show: { label: 'Faltou', variant: 'default' },
    // Generic statuses
    active: { label: 'Ativo', variant: 'success' },
    inactive: { label: 'Inativo', variant: 'default' },
    draft: { label: 'Rascunho', variant: 'default' },
    sent: { label: 'Enviado', variant: 'info' },
    approved: { label: 'Aprovado', variant: 'success' },
    rejected: { label: 'Rejeitado', variant: 'error' },
    expired: { label: 'Expirado', variant: 'warning' },
    in_progress: { label: 'Em Progresso', variant: 'info' },
    on_hold: { label: 'Em Espera', variant: 'warning' },
    todo: { label: 'Por Fazer', variant: 'default' },
    paid: { label: 'Pago', variant: 'success' },
    overdue: { label: 'Atrasado', variant: 'error' },
  }

  const allLabels = { ...defaultLabels, ...labels }
  const config = allLabels[status] || { label: status, variant: 'default' as BadgeVariant }

  return <Badge variant={config.variant}>{config.label}</Badge>
}
