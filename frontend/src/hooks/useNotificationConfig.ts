'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/Toast'
import { getErrorMessage } from '@/lib/api-error'
import type { NotificationConfig } from '@/types'

// Query keys factory
export const notificationConfigKeys = {
  all: ['notification-config'] as const,
  detail: () => [...notificationConfigKeys.all, 'detail'] as const,
}

interface NotificationConfigResponse {
  data: {
    config: NotificationConfig
  }
}

export function useNotificationConfig() {
  return useQuery({
    queryKey: notificationConfigKeys.detail(),
    queryFn: async () => {
      try {
        const response = await api.get<NotificationConfigResponse>('/notification-config')
        return response.data?.data?.config ?? null
      } catch {
        // Return null on error (e.g., module not enabled)
        return null
      }
    },
  })
}

interface UpdateNotificationConfigInput {
  whatsapp_enabled?: boolean
  twilio_account_sid?: string
  twilio_auth_token?: string
  twilio_whatsapp_number?: string
  reminder_24h_enabled?: boolean
  reminder_2h_enabled?: boolean
  reminder_24h_template?: string
  reminder_2h_template?: string
  confirmation_response_template?: string
}

export function useUpdateNotificationConfig() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (data: UpdateNotificationConfigInput) => {
      const response = await api.put<NotificationConfigResponse>('/notification-config', data)
      return response.data?.data?.config
    },
    // Optimistic update
    onMutate: async (data) => {
      await queryClient.cancelQueries({ queryKey: notificationConfigKeys.detail() })
      const previousConfig = queryClient.getQueryData<NotificationConfig>(
        notificationConfigKeys.detail()
      )

      if (previousConfig) {
        queryClient.setQueryData<NotificationConfig>(notificationConfigKeys.detail(), {
          ...previousConfig,
          ...data,
        })
      }

      return { previousConfig }
    },
    onError: (err, _, context) => {
      if (context?.previousConfig) {
        queryClient.setQueryData(notificationConfigKeys.detail(), context.previousConfig)
      }
      error('Erro ao atualizar configurações', getErrorMessage(err))
    },
    onSuccess: () => {
      success('Configurações atualizadas', 'As configurações de notificação foram guardadas.')
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: notificationConfigKeys.all })
    },
  })
}

interface TestWhatsAppInput {
  phone_number: string
  message?: string
}

interface TestWhatsAppResponse {
  data: {
    success: boolean
    message_sid?: string
    error?: string
  }
}

export function useTestWhatsApp() {
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (data: TestWhatsAppInput) => {
      const response = await api.post<TestWhatsAppResponse>('/notification-config/test', data)
      return response.data?.data
    },
    onSuccess: (result) => {
      if (result?.success) {
        success('Mensagem enviada', 'A mensagem de teste foi enviada com sucesso.')
      } else {
        error('Erro no envio', result?.error || 'Não foi possível enviar a mensagem.')
      }
    },
    onError: (err) => {
      error('Erro ao testar WhatsApp', getErrorMessage(err))
    },
  })
}
