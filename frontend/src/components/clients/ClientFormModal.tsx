'use client'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Modal } from '@/components/ui/Modal'
import { FormField, FormTextarea, FormError, SubmitButton } from '@/components/forms/FormField'
import { createClientSchema, type CreateClientFormData } from '@/schemas/client'
import { useCreateClient, useUpdateClient } from '@/hooks/useClients'
import { getErrorMessage } from '@/lib/errors'
import type { Client } from '@/types'

interface ClientFormModalProps {
  client: Client | null
  isOpen: boolean
  onClose: () => void
  onSuccess?: () => void
}

export function ClientFormModal({ client, isOpen, onClose, onSuccess }: ClientFormModalProps) {
  const createMutation = useCreateClient()
  const updateMutation = useUpdateClient()
  const isEditing = !!client

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<CreateClientFormData>({
    resolver: zodResolver(createClientSchema),
    defaultValues: {
      name: '',
      email: '',
      phone: '',
      address: '',
      notes: '',
    },
  })

  // Reset form when client changes or modal opens
  useEffect(() => {
    if (isOpen) {
      reset({
        name: client?.name || '',
        email: client?.email || '',
        phone: client?.phone || '',
        address: client?.address || '',
        notes: client?.notes || '',
      })
    }
  }, [client, isOpen, reset])

  const onSubmit = async (data: CreateClientFormData) => {
    try {
      if (isEditing) {
        await updateMutation.mutateAsync({ id: client.id, data })
      } else {
        await createMutation.mutateAsync(data)
      }
      onSuccess?.()
      onClose()
    } catch {
      // Error is handled by the mutation
    }
  }

  const mutation = isEditing ? updateMutation : createMutation

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={isEditing ? 'Editar Cliente' : 'Novo Cliente'}
      size="lg"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <FormError message={mutation.error ? getErrorMessage(mutation.error) : undefined} />

        <FormField
          id="name"
          type="text"
          label="Nome"
          required
          error={errors.name?.message}
          {...register('name')}
        />

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <FormField
            id="email"
            type="email"
            label="Email"
            required
            error={errors.email?.message}
            {...register('email')}
          />

          <FormField
            id="phone"
            type="tel"
            label="Telefone"
            required
            error={errors.phone?.message}
            {...register('phone')}
          />
        </div>

        <FormField
          id="address"
          type="text"
          label="Morada"
          error={errors.address?.message}
          {...register('address')}
        />

        <FormTextarea
          id="notes"
          label="Notas"
          rows={3}
          error={errors.notes?.message}
          {...register('notes')}
        />

        <div className="flex justify-end space-x-3 pt-4">
          <button type="button" onClick={onClose} className="btn btn-secondary">
            Cancelar
          </button>
          <SubmitButton
            isLoading={isSubmitting || mutation.isPending}
            loadingText="A guardar..."
          >
            Guardar
          </SubmitButton>
        </div>
      </form>
    </Modal>
  )
}
