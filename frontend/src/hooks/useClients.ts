import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { Client } from '@/types'
import type { CreateClientFormData, UpdateClientFormData } from '@/schemas/client'

const CLIENTS_QUERY_KEY = ['clients'] as const

interface ClientsParams {
  page?: number
  limit?: number
  search?: string
}

interface ClientsResponse {
  data: Client[]
  pagination: {
    page: number
    limit: number
    total: number
    total_pages: number
  }
}

export function useClients(params?: ClientsParams) {
  return useQuery({
    queryKey: [...CLIENTS_QUERY_KEY, params],
    queryFn: async (): Promise<ClientsResponse> => {
      const searchParams = new URLSearchParams()
      if (params?.page) searchParams.set('page', params.page.toString())
      if (params?.limit) searchParams.set('limit', params.limit.toString())
      if (params?.search) searchParams.set('search', params.search)

      const queryString = searchParams.toString()
      const url = queryString ? `/clients?${queryString}` : '/clients'
      const response = await api.get<{ data: { clients: Client[], total: number, page: number, limit: number } }>(url)

      const { clients, total, page, limit } = response.data.data
      return {
        data: clients || [],
        pagination: {
          page: page || 1,
          limit: limit || 20,
          total: total || 0,
          total_pages: Math.ceil((total || 0) / (limit || 20)),
        },
      }
    },
    staleTime: 60 * 1000, // 1 minute
  })
}

export function useClient(id: string) {
  return useQuery({
    queryKey: [...CLIENTS_QUERY_KEY, id],
    queryFn: () => api.getClients().then((clients) => clients.find((c) => c.id === id)),
    enabled: !!id,
  })
}

export function useCreateClient() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateClientFormData) => api.createClient(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CLIENTS_QUERY_KEY })
    },
  })
}

export function useUpdateClient() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateClientFormData }) =>
      api.updateClient(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CLIENTS_QUERY_KEY })
    },
  })
}

export function useDeleteClient() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => api.deleteClient(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CLIENTS_QUERY_KEY })
    },
  })
}

// Search clients with memoization
export function useFilteredClients(clients: Client[] | undefined, searchQuery: string) {
  if (!clients) return []
  if (!searchQuery) return clients

  const query = searchQuery.toLowerCase()
  return clients.filter(
    (client) =>
      client.name.toLowerCase().includes(query) ||
      client.email.toLowerCase().includes(query) ||
      client.phone?.includes(query)
  )
}
