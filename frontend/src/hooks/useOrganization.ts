'use client'

import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Organization } from '@/types'

export const organizationKeys = {
  all: ['organization'] as const,
  current: () => [...organizationKeys.all, 'current'] as const,
}

interface OrganizationResponse {
  data: Organization
}

export function useOrganization() {
  return useQuery({
    queryKey: organizationKeys.current(),
    queryFn: async () => {
      const response = await api.get<OrganizationResponse>('/organizations')
      return response.data.data
    },
  })
}
