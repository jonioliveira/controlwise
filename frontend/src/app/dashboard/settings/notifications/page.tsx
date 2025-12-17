'use client'

import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import Link from 'next/link'
import {
  ArrowLeft,
  Bell,
  MessageCircle,
  CheckCircle,
  AlertCircle,
  Loader2,
  Send,
  Info
} from 'lucide-react'
import {
  useNotificationConfig,
  useUpdateNotificationConfig,
  useTestWhatsApp
} from '@/hooks/useNotificationConfig'

interface NotificationFormData {
  whatsapp_enabled: boolean
  twilio_account_sid: string
  twilio_auth_token: string
  twilio_whatsapp_number: string
  reminder_24h_enabled: boolean
  reminder_2h_enabled: boolean
  reminder_24h_template: string
  reminder_2h_template: string
  confirmation_response_template: string
}

const defaultTemplates = {
  reminder_24h: `Olá {{patient_name}}, este é um lembrete da sua consulta amanhã às {{time}} com {{therapist_name}}.

Responda:
1 - Confirmar
2 - Cancelar

Obrigado!`,
  reminder_2h: `Olá {{patient_name}}, a sua consulta com {{therapist_name}} é daqui a 2 horas ({{time}}).

Aguardamos por si!`,
  confirmation_response: `Obrigado pela sua resposta! A sua consulta foi {{status}}.`,
}

export default function NotificationsSettingsPage() {
  const { data: config, isLoading, error } = useNotificationConfig()
  const updateConfig = useUpdateNotificationConfig()
  const testWhatsApp = useTestWhatsApp()

  const [testPhone, setTestPhone] = useState('')
  const [testResult, setTestResult] = useState<{
    success: boolean
    message: string
  } | null>(null)

  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { isDirty, isSubmitting },
  } = useForm<NotificationFormData>({
    defaultValues: {
      whatsapp_enabled: false,
      twilio_account_sid: '',
      twilio_auth_token: '',
      twilio_whatsapp_number: '',
      reminder_24h_enabled: true,
      reminder_2h_enabled: true,
      reminder_24h_template: defaultTemplates.reminder_24h,
      reminder_2h_template: defaultTemplates.reminder_2h,
      confirmation_response_template: defaultTemplates.confirmation_response,
    },
  })

  const whatsappEnabled = watch('whatsapp_enabled')

  useEffect(() => {
    if (config) {
      reset({
        whatsapp_enabled: config.whatsapp_enabled,
        twilio_account_sid: '', // Don't populate sensitive data
        twilio_auth_token: '',
        twilio_whatsapp_number: config.twilio_whatsapp_number || '',
        reminder_24h_enabled: config.reminder_24h_enabled,
        reminder_2h_enabled: config.reminder_2h_enabled,
        reminder_24h_template: config.reminder_24h_template || defaultTemplates.reminder_24h,
        reminder_2h_template: config.reminder_2h_template || defaultTemplates.reminder_2h,
        confirmation_response_template: config.confirmation_response_template || defaultTemplates.confirmation_response,
      })
    }
  }, [config, reset])

  const onSubmit = async (data: NotificationFormData) => {
    try {
      // Only include non-empty values
      const payload: Record<string, unknown> = {
        whatsapp_enabled: data.whatsapp_enabled,
        reminder_24h_enabled: data.reminder_24h_enabled,
        reminder_2h_enabled: data.reminder_2h_enabled,
        reminder_24h_template: data.reminder_24h_template,
        reminder_2h_template: data.reminder_2h_template,
        confirmation_response_template: data.confirmation_response_template,
      }

      if (data.twilio_account_sid) {
        payload.twilio_account_sid = data.twilio_account_sid
      }
      if (data.twilio_auth_token) {
        payload.twilio_auth_token = data.twilio_auth_token
      }
      if (data.twilio_whatsapp_number) {
        payload.twilio_whatsapp_number = data.twilio_whatsapp_number
      }

      await updateConfig.mutateAsync(payload)
    } catch (error) {
      console.error('Failed to update config:', error)
    }
  }

  const handleTestWhatsApp = async () => {
    if (!testPhone) return
    setTestResult(null)
    try {
      const result = await testWhatsApp.mutateAsync({
        phone_number: testPhone,
        message: 'Teste de configuração WhatsApp do ControlWise.',
      })
      setTestResult({
        success: result.success,
        message: result.success
          ? 'Mensagem enviada com sucesso!'
          : result.error || 'Falha ao enviar mensagem',
      })
    } catch (error) {
      setTestResult({
        success: false,
        message: 'Erro ao enviar mensagem de teste',
      })
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-md bg-red-50 p-4">
        <p className="text-sm text-red-700">
          Erro ao carregar configurações. Por favor tente novamente.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          href="/dashboard/settings"
          className="flex items-center justify-center h-10 w-10 rounded-lg border border-gray-200 bg-white text-gray-500 hover:bg-gray-50 hover:text-gray-700"
        >
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Notificações</h1>
          <p className="mt-1 text-sm text-gray-500">
            Configurar WhatsApp e lembretes automáticos
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* WhatsApp Configuration */}
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="p-6 border-b border-gray-200">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100">
                <MessageCircle className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-900">WhatsApp</h2>
                <p className="text-sm text-gray-500">
                  Configuração da integração com Twilio WhatsApp
                </p>
              </div>
            </div>
          </div>

          <div className="p-6 space-y-4">
            {/* Enable toggle */}
            <div className="flex items-center justify-between">
              <div>
                <label className="text-sm font-medium text-gray-900">
                  Ativar WhatsApp
                </label>
                <p className="text-sm text-gray-500">
                  Enviar lembretes e confirmações por WhatsApp
                </p>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  {...register('whatsapp_enabled')}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
              </label>
            </div>

            {/* Twilio status */}
            {config?.twilio_configured ? (
              <div className="flex items-center gap-2 text-sm text-green-600 bg-green-50 p-3 rounded-lg">
                <CheckCircle className="h-4 w-4" />
                Twilio configurado
              </div>
            ) : (
              <div className="flex items-center gap-2 text-sm text-amber-600 bg-amber-50 p-3 rounded-lg">
                <AlertCircle className="h-4 w-4" />
                Twilio não configurado - configure as credenciais abaixo
              </div>
            )}

            {/* Twilio credentials */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Twilio Account SID
                </label>
                <input
                  type="text"
                  {...register('twilio_account_sid')}
                  placeholder={config?.twilio_configured ? '••••••••••' : 'ACxxxxxxxxxx'}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Twilio Auth Token
                </label>
                <input
                  type="password"
                  {...register('twilio_auth_token')}
                  placeholder={config?.twilio_configured ? '••••••••••' : 'Auth token'}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Número WhatsApp Twilio
                </label>
                <input
                  type="text"
                  {...register('twilio_whatsapp_number')}
                  placeholder="whatsapp:+351912345678"
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Formato: whatsapp:+[código país][número]
                </p>
              </div>
            </div>

            {/* Test WhatsApp */}
            {config?.twilio_configured && whatsappEnabled && (
              <div className="border-t border-gray-200 pt-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Testar Envio
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={testPhone}
                    onChange={(e) => setTestPhone(e.target.value)}
                    placeholder="+351912345678"
                    className="flex-1 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
                  />
                  <button
                    type="button"
                    onClick={handleTestWhatsApp}
                    disabled={!testPhone || testWhatsApp.isPending}
                    className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-md hover:bg-green-700 disabled:opacity-50"
                  >
                    {testWhatsApp.isPending ? (
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    ) : (
                      <Send className="h-4 w-4 mr-2" />
                    )}
                    Enviar Teste
                  </button>
                </div>
                {testResult && (
                  <div
                    className={`mt-2 flex items-center gap-2 text-sm p-2 rounded ${
                      testResult.success
                        ? 'text-green-600 bg-green-50'
                        : 'text-red-600 bg-red-50'
                    }`}
                  >
                    {testResult.success ? (
                      <CheckCircle className="h-4 w-4" />
                    ) : (
                      <AlertCircle className="h-4 w-4" />
                    )}
                    {testResult.message}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Reminder Settings */}
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="p-6 border-b border-gray-200">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
                <Bell className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-900">Lembretes</h2>
                <p className="text-sm text-gray-500">
                  Configurar lembretes automáticos para sessões
                </p>
              </div>
            </div>
          </div>

          <div className="p-6 space-y-6">
            {/* 24h Reminder */}
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <label className="text-sm font-medium text-gray-900">
                    Lembrete 24 horas antes
                  </label>
                  <p className="text-sm text-gray-500">
                    Enviar lembrete um dia antes da sessão
                  </p>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    {...register('reminder_24h_enabled')}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                </label>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Modelo da Mensagem (24h)
                </label>
                <textarea
                  {...register('reminder_24h_template')}
                  rows={5}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Variáveis: {'{{patient_name}}'}, {'{{therapist_name}}'}, {'{{date}}'}, {'{{time}}'}
                </p>
              </div>
            </div>

            {/* 2h Reminder */}
            <div className="space-y-4 border-t border-gray-200 pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <label className="text-sm font-medium text-gray-900">
                    Lembrete 2 horas antes
                  </label>
                  <p className="text-sm text-gray-500">
                    Enviar lembrete 2 horas antes da sessão
                  </p>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    {...register('reminder_2h_enabled')}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
                </label>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Modelo da Mensagem (2h)
                </label>
                <textarea
                  {...register('reminder_2h_template')}
                  rows={4}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                />
              </div>
            </div>

            {/* Confirmation Response */}
            <div className="space-y-4 border-t border-gray-200 pt-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Modelo de Resposta à Confirmação
                </label>
                <textarea
                  {...register('confirmation_response_template')}
                  rows={3}
                  className="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-sm"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Variáveis: {'{{status}}'} (confirmada/cancelada)
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Save Button */}
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={!isDirty || isSubmitting}
            className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-primary-600 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50"
          >
            {isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
            Guardar Alterações
          </button>
        </div>
      </form>

      {/* Info Box */}
      <div className="bg-blue-50 rounded-lg p-4 flex items-start gap-3">
        <Info className="h-5 w-5 text-blue-600 shrink-0 mt-0.5" />
        <div className="text-sm text-blue-800">
          <p className="font-medium">Sobre a integração WhatsApp</p>
          <p className="mt-1">
            Os lembretes são enviados automaticamente através do Twilio WhatsApp Business API.
            É necessário ter uma conta Twilio com WhatsApp Business API configurada.
          </p>
          <a
            href="https://www.twilio.com/whatsapp"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center mt-2 text-blue-700 hover:underline"
          >
            Saber mais sobre Twilio WhatsApp
          </a>
        </div>
      </div>
    </div>
  )
}
