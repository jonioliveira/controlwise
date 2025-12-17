'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useToast } from '@/components/ui/Toast'
import { getErrorMessage } from '@/lib/api-error'
import type { OrganizationModule, ModuleName } from '@/types'

// Query keys factory
export const moduleKeys = {
  all: ['modules'] as const,
  lists: () => [...moduleKeys.all, 'list'] as const,
  enabled: () => [...moduleKeys.all, 'enabled'] as const,
}

interface ModulesResponse {
  data: {
    modules: OrganizationModule[]
  }
}

interface EnabledModulesResponse {
  data: {
    enabled_modules: ModuleName[]
  }
}

export function useModules() {
  return useQuery({
    queryKey: moduleKeys.lists(),
    queryFn: async () => {
      try {
        const response = await api.get<ModulesResponse>('/modules')
        return response.data?.data?.modules ?? []
      } catch {
        // Return empty array on error to avoid undefined
        return []
      }
    },
  })
}

export function useEnabledModules() {
  return useQuery({
    queryKey: moduleKeys.enabled(),
    queryFn: async () => {
      try {
        const response = await api.get<EnabledModulesResponse>('/modules/enabled')
        return response.data?.data?.enabled_modules ?? []
      } catch {
        // Return empty array on error to avoid undefined
        return []
      }
    },
  })
}

export function useModuleEnabled(moduleName: ModuleName) {
  const { data: enabledModules } = useEnabledModules()
  return enabledModules?.includes(moduleName) ?? false
}

export function useEnableModule() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (moduleName: ModuleName) => {
      const response = await api.post(`/modules/${moduleName}/enable`)
      return response.data
    },
    // Optimistic update
    onMutate: async (moduleName) => {
      await queryClient.cancelQueries({ queryKey: moduleKeys.enabled() })
      const previousEnabled = queryClient.getQueryData<ModuleName[]>(moduleKeys.enabled())

      if (previousEnabled && !previousEnabled.includes(moduleName)) {
        queryClient.setQueryData<ModuleName[]>(
          moduleKeys.enabled(),
          [...previousEnabled, moduleName]
        )
      }

      return { previousEnabled }
    },
    onError: (err, _, context) => {
      if (context?.previousEnabled) {
        queryClient.setQueryData(moduleKeys.enabled(), context.previousEnabled)
      }
      error('Erro ao ativar módulo', getErrorMessage(err))
    },
    onSuccess: () => {
      success('Módulo ativado', 'O módulo foi ativado com sucesso.')
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: moduleKeys.all })
    },
  })
}

export function useDisableModule() {
  const queryClient = useQueryClient()
  const { success, error } = useToast()

  return useMutation({
    mutationFn: async (moduleName: ModuleName) => {
      const response = await api.post(`/modules/${moduleName}/disable`)
      return response.data
    },
    // Optimistic update
    onMutate: async (moduleName) => {
      await queryClient.cancelQueries({ queryKey: moduleKeys.enabled() })
      const previousEnabled = queryClient.getQueryData<ModuleName[]>(moduleKeys.enabled())

      if (previousEnabled) {
        queryClient.setQueryData<ModuleName[]>(
          moduleKeys.enabled(),
          previousEnabled.filter((m) => m !== moduleName)
        )
      }

      return { previousEnabled }
    },
    onError: (err, _, context) => {
      if (context?.previousEnabled) {
        queryClient.setQueryData(moduleKeys.enabled(), context.previousEnabled)
      }
      error('Erro ao desativar módulo', getErrorMessage(err))
    },
    onSuccess: () => {
      success('Módulo desativado', 'O módulo foi desativado com sucesso.')
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: moduleKeys.all })
    },
  })
}

// Helper function to filter enabled modules from the list
export function useFilteredModules(modules: OrganizationModule[] | undefined) {
  if (!modules) return []
  return modules.filter((m) => m.is_enabled)
}
