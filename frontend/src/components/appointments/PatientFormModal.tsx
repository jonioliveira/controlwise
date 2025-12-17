'use client'

import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Modal } from '@/components/ui/Modal'
import { patientUpdateSchema, type PatientUpdateFormData } from '@/schemas/patient'
import { useCreatePatient, useUpdatePatient, usePatient } from '@/hooks/usePatients'
import { useClients, useCreateClient } from '@/hooks/useClients'
import { useEnabledModules } from '@/hooks/useModules'
import { Loader2, User, Mail, Phone, Search, UserPlus, Users } from 'lucide-react'

interface PatientFormModalProps {
  isOpen: boolean
  onClose: () => void
  patientId?: string | null
}

// Schema for creating patient with existing client
const patientWithExistingClientSchema = z.object({
  client_id: z.string().min(1, 'Cliente é obrigatório').uuid('ID de cliente inválido'),
  date_of_birth: z.string().optional(),
  notes: z.string().max(5000).optional(),
  emergency_contact: z.string().max(200).optional(),
  emergency_phone: z.string().max(20).regex(/^[+]?[0-9\s-]*$/, 'Telefone inválido').optional().or(z.literal('')),
})

// Schema for creating patient with new client
const patientWithNewClientSchema = z.object({
  // Client fields
  client_name: z.string().min(2, 'Nome deve ter pelo menos 2 caracteres').max(200),
  client_email: z.string().email('Email inválido').max(255),
  client_phone: z.string().min(9, 'Telefone deve ter pelo menos 9 dígitos').max(20).regex(/^[+]?[0-9\s-]+$/, 'Telefone inválido'),
  // Patient fields
  date_of_birth: z.string().optional(),
  notes: z.string().max(5000).optional(),
  emergency_contact: z.string().max(200).optional(),
  emergency_phone: z.string().max(20).regex(/^[+]?[0-9\s-]*$/, 'Telefone inválido').optional().or(z.literal('')),
})

type PatientWithExistingClientData = z.infer<typeof patientWithExistingClientSchema>
type PatientWithNewClientData = z.infer<typeof patientWithNewClientSchema>

export function PatientFormModal({ isOpen, onClose, patientId }: PatientFormModalProps) {
  const isEditing = !!patientId
  const { data: patient, isLoading: isLoadingPatient } = usePatient(patientId ?? null)
  const { data: clientsData, isLoading: isLoadingClients } = useClients({ limit: 100 })
  const { data: enabledModules = [] } = useEnabledModules()
  const createPatient = useCreatePatient()
  const updatePatient = useUpdatePatient()
  const createClient = useCreateClient()

  // Check if we're in pure healthcare mode (appointments without construction)
  // In this mode, new patient = new client, so hide the toggle
  const hasConstruction = enabledModules.includes('construction')
  const hasAppointments = enabledModules.includes('appointments')
  const isPureHealthcareMode = hasAppointments && !hasConstruction

  // Toggle: false = search existing client, true = create new client
  // In pure healthcare mode, always create new client
  const [createNewClient, setCreateNewClient] = useState(isPureHealthcareMode)
  const [clientSearch, setClientSearch] = useState('')
  const [showClientDropdown, setShowClientDropdown] = useState(false)
  const [selectedClient, setSelectedClient] = useState<{ id: string; name: string; email: string; phone: string } | null>(null)

  // Filter clients based on search
  const filteredClients = clientsData?.data?.filter((client) => {
    if (!clientSearch) return true
    const search = clientSearch.toLowerCase()
    return (
      client.name.toLowerCase().includes(search) ||
      client.email.toLowerCase().includes(search) ||
      client.phone?.toLowerCase().includes(search)
    )
  }) ?? []

  // Form for existing client mode
  const existingClientForm = useForm<PatientWithExistingClientData>({
    resolver: zodResolver(patientWithExistingClientSchema),
    defaultValues: {
      client_id: '',
      date_of_birth: '',
      notes: '',
      emergency_contact: '',
      emergency_phone: '',
    },
  })

  // Form for new client mode
  const newClientForm = useForm<PatientWithNewClientData>({
    resolver: zodResolver(patientWithNewClientSchema),
    defaultValues: {
      client_name: '',
      client_email: '',
      client_phone: '',
      date_of_birth: '',
      notes: '',
      emergency_contact: '',
      emergency_phone: '',
    },
  })

  // Form for editing (update mode)
  const updateForm = useForm<PatientUpdateFormData>({
    resolver: zodResolver(patientUpdateSchema),
    defaultValues: {
      date_of_birth: '',
      notes: '',
      emergency_contact: '',
      emergency_phone: '',
    },
  })

  // Reset forms when modal opens/closes or mode changes
  useEffect(() => {
    if (patient && isEditing) {
      updateForm.reset({
        date_of_birth: patient.date_of_birth || '',
        notes: patient.notes || '',
        emergency_contact: patient.emergency_contact || '',
        emergency_phone: patient.emergency_phone || '',
      })
      setSelectedClient({
        id: patient.client_id,
        name: patient.client_name,
        email: patient.client_email,
        phone: patient.client_phone,
      })
    } else if (!isEditing) {
      existingClientForm.reset()
      newClientForm.reset()
      setSelectedClient(null)
      setClientSearch('')
      // In pure healthcare mode, always default to new client
      setCreateNewClient(isPureHealthcareMode)
    }
  }, [patient, isEditing, isOpen, isPureHealthcareMode])

  const handleClientSelect = (client: { id: string; name: string; email: string; phone: string }) => {
    setSelectedClient(client)
    existingClientForm.setValue('client_id', client.id)
    setShowClientDropdown(false)
    setClientSearch('')
  }

  const handleModeChange = (newMode: boolean) => {
    setCreateNewClient(newMode)
    setSelectedClient(null)
    existingClientForm.reset()
    newClientForm.reset()
    setClientSearch('')
  }

  const onSubmitExistingClient = async (data: PatientWithExistingClientData) => {
    try {
      const cleanedData = {
        client_id: data.client_id,
        date_of_birth: data.date_of_birth || undefined,
        notes: data.notes || undefined,
        emergency_contact: data.emergency_contact || undefined,
        emergency_phone: data.emergency_phone || undefined,
      }
      await createPatient.mutateAsync(cleanedData)
      onClose()
    } catch (error) {
      console.error('Failed to save patient:', error)
    }
  }

  const onSubmitNewClient = async (data: PatientWithNewClientData) => {
    try {
      // First create the client
      const newClient = await createClient.mutateAsync({
        name: data.client_name,
        email: data.client_email,
        phone: data.client_phone,
      })

      // Then create the patient linked to the new client
      const patientData = {
        client_id: newClient.id,
        date_of_birth: data.date_of_birth || undefined,
        notes: data.notes || undefined,
        emergency_contact: data.emergency_contact || undefined,
        emergency_phone: data.emergency_phone || undefined,
      }
      await createPatient.mutateAsync(patientData)
      onClose()
    } catch (error) {
      console.error('Failed to save patient:', error)
    }
  }

  const onSubmitUpdate = async (data: PatientUpdateFormData) => {
    if (!patientId) return
    try {
      const cleanedData = {
        date_of_birth: data.date_of_birth || undefined,
        notes: data.notes || undefined,
        emergency_contact: data.emergency_contact || undefined,
        emergency_phone: data.emergency_phone || undefined,
        is_active: data.is_active,
      }
      await updatePatient.mutateAsync({ id: patientId, data: cleanedData })
      onClose()
    } catch (error) {
      console.error('Failed to update patient:', error)
    }
  }

  if (isEditing && isLoadingPatient) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} title="Carregar Paciente">
        <div className="flex items-center justify-center py-8">
          <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
        </div>
      </Modal>
    )
  }

  // Editing mode
  if (isEditing) {
    return (
      <Modal isOpen={isOpen} onClose={onClose} title="Editar Paciente">
        <form onSubmit={updateForm.handleSubmit(onSubmitUpdate)} className="space-y-4">
          {/* Client Info Display (read-only) */}
          {selectedClient && (
            <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg">
              <h3 className="text-sm font-medium text-gray-700 mb-3">Dados do Cliente</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-gray-400" />
                  <span className="text-sm text-gray-900">{selectedClient.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Mail className="h-4 w-4 text-gray-400" />
                  <span className="text-sm text-gray-900">{selectedClient.email}</span>
                </div>
                <div className="flex items-center gap-2">
                  <Phone className="h-4 w-4 text-gray-400" />
                  <span className="text-sm text-gray-900">{selectedClient.phone}</span>
                </div>
              </div>
              <p className="text-xs text-gray-500 mt-2">
                Para alterar estes dados, edite o cliente diretamente.
              </p>
            </div>
          )}

          {/* Healthcare Fields */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Data de Nascimento
              </label>
              <input
                type="date"
                {...updateForm.register('date_of_birth')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Contacto de Emergência
              </label>
              <input
                type="text"
                {...updateForm.register('emergency_contact')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="Nome do contacto"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Telefone de Emergência
              </label>
              <input
                type="tel"
                {...updateForm.register('emergency_phone')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="+351 912 345 678"
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Notas Médicas
              </label>
              <textarea
                {...updateForm.register('notes')}
                rows={3}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="Observações, alergias, histórico médico..."
              />
            </div>
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
              disabled={updateForm.formState.isSubmitting}
              className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 disabled:opacity-50"
            >
              {updateForm.formState.isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Guardar
            </button>
          </div>
        </form>
      </Modal>
    )
  }

  // Creation mode
  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Novo Paciente">
      {/* Toggle Switch - Only show when not in pure healthcare mode */}
      {/* In pure healthcare mode, new patient = new client always */}
      {!isPureHealthcareMode && (
        <div className="mb-6">
          <div className="flex items-center justify-center p-1 bg-gray-100 rounded-lg">
            <button
              type="button"
              onClick={() => handleModeChange(false)}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                !createNewClient
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              <Users className="h-4 w-4" />
              Cliente Existente
            </button>
            <button
              type="button"
              onClick={() => handleModeChange(true)}
              className={`flex-1 flex items-center justify-center gap-2 px-4 py-2 text-sm font-medium rounded-md transition-colors ${
                createNewClient
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              <UserPlus className="h-4 w-4" />
              Novo Cliente
            </button>
          </div>
        </div>
      )}

      {/* Existing Client Mode */}
      {!createNewClient && (
        <form onSubmit={existingClientForm.handleSubmit(onSubmitExistingClient)} className="space-y-4">
          {/* Client Selection */}
          <div className="relative">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Cliente *
            </label>
            {selectedClient ? (
              <div className="p-3 bg-gray-50 border border-gray-200 rounded-lg">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-primary-100 rounded-full flex items-center justify-center">
                      <span className="text-primary-700 font-medium">
                        {selectedClient.name.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">{selectedClient.name}</p>
                      <p className="text-sm text-gray-500">{selectedClient.email}</p>
                    </div>
                  </div>
                  <button
                    type="button"
                    onClick={() => {
                      setSelectedClient(null)
                      existingClientForm.setValue('client_id', '')
                    }}
                    className="text-sm text-primary-600 hover:text-primary-700"
                  >
                    Alterar
                  </button>
                </div>
              </div>
            ) : (
              <div className="relative">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <input
                    type="text"
                    value={clientSearch}
                    onChange={(e) => {
                      setClientSearch(e.target.value)
                      setShowClientDropdown(true)
                    }}
                    onFocus={() => setShowClientDropdown(true)}
                    className="block w-full pl-10 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                    placeholder="Pesquisar cliente por nome, email ou telefone..."
                  />
                </div>
                {showClientDropdown && (
                  <div className="absolute z-10 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                    {isLoadingClients ? (
                      <div className="p-3 text-center text-gray-500">
                        <Loader2 className="h-4 w-4 animate-spin inline mr-2" />
                        A carregar clientes...
                      </div>
                    ) : filteredClients.length === 0 ? (
                      <div className="p-3 text-center text-gray-500">
                        Nenhum cliente encontrado
                      </div>
                    ) : (
                      filteredClients.map((client) => (
                        <button
                          key={client.id}
                          type="button"
                          onClick={() => handleClientSelect(client)}
                          className="w-full p-3 text-left hover:bg-gray-50 flex items-center gap-3 border-b border-gray-100 last:border-0"
                        >
                          <div className="w-8 h-8 bg-primary-100 rounded-full flex items-center justify-center flex-shrink-0">
                            <span className="text-primary-700 text-sm font-medium">
                              {client.name.charAt(0).toUpperCase()}
                            </span>
                          </div>
                          <div className="min-w-0">
                            <p className="font-medium text-gray-900 truncate">{client.name}</p>
                            <p className="text-xs text-gray-500 truncate">{client.email} - {client.phone}</p>
                          </div>
                        </button>
                      ))
                    )}
                  </div>
                )}
              </div>
            )}
            <input type="hidden" {...existingClientForm.register('client_id')} />
            {existingClientForm.formState.errors.client_id && (
              <p className="mt-1 text-sm text-red-600">{existingClientForm.formState.errors.client_id.message}</p>
            )}
          </div>

          {/* Healthcare Fields */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Data de Nascimento
              </label>
              <input
                type="date"
                {...existingClientForm.register('date_of_birth')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Contacto de Emergência
              </label>
              <input
                type="text"
                {...existingClientForm.register('emergency_contact')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="Nome do contacto"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Telefone de Emergência
              </label>
              <input
                type="tel"
                {...existingClientForm.register('emergency_phone')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="+351 912 345 678"
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Notas Médicas
              </label>
              <textarea
                {...existingClientForm.register('notes')}
                rows={3}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="Observações, alergias, histórico médico..."
              />
            </div>
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
              disabled={existingClientForm.formState.isSubmitting || !selectedClient}
              className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 disabled:opacity-50"
            >
              {existingClientForm.formState.isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Criar Paciente
            </button>
          </div>
        </form>
      )}

      {/* New Client Mode */}
      {createNewClient && (
        <form onSubmit={newClientForm.handleSubmit(onSubmitNewClient)} className="space-y-4">
          {/* Client Creation Fields */}
          <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <h3 className="text-sm font-medium text-blue-800 mb-3 flex items-center gap-2">
              <UserPlus className="h-4 w-4" />
              Dados do Novo Cliente
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Nome *
                </label>
                <input
                  type="text"
                  {...newClientForm.register('client_name')}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  placeholder="Nome completo"
                />
                {newClientForm.formState.errors.client_name && (
                  <p className="mt-1 text-sm text-red-600">{newClientForm.formState.errors.client_name.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email *
                </label>
                <input
                  type="email"
                  {...newClientForm.register('client_email')}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  placeholder="email@exemplo.com"
                />
                {newClientForm.formState.errors.client_email && (
                  <p className="mt-1 text-sm text-red-600">{newClientForm.formState.errors.client_email.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Telefone *
                </label>
                <input
                  type="tel"
                  {...newClientForm.register('client_phone')}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  placeholder="+351 912 345 678"
                />
                {newClientForm.formState.errors.client_phone && (
                  <p className="mt-1 text-sm text-red-600">{newClientForm.formState.errors.client_phone.message}</p>
                )}
              </div>
            </div>
          </div>

          {/* Healthcare Fields */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Data de Nascimento
              </label>
              <input
                type="date"
                {...newClientForm.register('date_of_birth')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Contacto de Emergência
              </label>
              <input
                type="text"
                {...newClientForm.register('emergency_contact')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="Nome do contacto"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Telefone de Emergência
              </label>
              <input
                type="tel"
                {...newClientForm.register('emergency_phone')}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="+351 912 345 678"
              />
            </div>

            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Notas Médicas
              </label>
              <textarea
                {...newClientForm.register('notes')}
                rows={3}
                className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                placeholder="Observações, alergias, histórico médico..."
              />
            </div>
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
              disabled={newClientForm.formState.isSubmitting || createClient.isPending}
              className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 disabled:opacity-50"
            >
              {(newClientForm.formState.isSubmitting || createClient.isPending) && (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              )}
              Criar Cliente e Paciente
            </button>
          </div>
        </form>
      )}
    </Modal>
  )
}
