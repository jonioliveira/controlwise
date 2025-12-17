import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/authStore'
import type { LoginFormData, RegisterFormData } from '@/schemas/auth'

export function useLogin() {
  const router = useRouter()
  const setAuth = useAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (data: LoginFormData) => {
      const response = await api.login(data)
      return response
    },
    onSuccess: (data) => {
      setAuth(data.user, data.token)
      router.push('/dashboard')
    },
  })
}

export function useRegister() {
  const router = useRouter()
  const setAuth = useAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (data: Omit<RegisterFormData, 'confirmPassword'>) => {
      const response = await api.register(data)
      return response
    },
    onSuccess: (data) => {
      setAuth(data.user, data.token)
      router.push('/dashboard')
    },
  })
}

export function useLogout() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const logout = useAuthStore((state) => state.logout)

  return useMutation({
    mutationFn: async () => {
      await api.logout()
    },
    onSuccess: () => {
      logout()
      queryClient.clear()
      router.push('/login')
    },
    onError: () => {
      // Even if the API call fails, log out locally
      logout()
      queryClient.clear()
      router.push('/login')
    },
  })
}

export function useCurrentUser() {
  const { isAuthenticated, setUser, setLoading } = useAuthStore()

  return useQuery({
    queryKey: ['currentUser'],
    queryFn: async () => {
      const user = await api.getCurrentUser()
      setUser(user)
      return user
    },
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: false,
    meta: {
      onSettled: () => {
        setLoading(false)
      },
    },
  })
}
