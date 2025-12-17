import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { SystemAdmin, ImpersonationToken, User } from '@/types'
import { adminApi } from '@/lib/adminApi'

interface AdminAuthState {
  admin: SystemAdmin | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  // Impersonation state
  isImpersonating: boolean
  impersonationToken: string | null
  impersonatedUser: User | null
  impersonationSessionId: string | null
  originalAdminToken: string | null

  // Actions
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  loadAdmin: () => Promise<void>
  setToken: (token: string) => void

  // Impersonation actions
  startImpersonation: (userId: string, reason: string) => Promise<void>
  endImpersonation: () => Promise<void>
}

export const useAdminAuthStore = create<AdminAuthState>()(
  persist(
    (set, get) => ({
      admin: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      isImpersonating: false,
      impersonationToken: null,
      impersonatedUser: null,
      impersonationSessionId: null,
      originalAdminToken: null,

      login: async (email: string, password: string) => {
        set({ isLoading: true })
        try {
          const response = await adminApi.login({ email, password })
          adminApi.setToken(response.token)
          set({
            admin: response.admin,
            token: response.token,
            isAuthenticated: true,
            isLoading: false,
          })
        } catch (error) {
          set({ isLoading: false })
          throw error
        }
      },

      logout: () => {
        adminApi.setToken(null)
        set({
          admin: null,
          token: null,
          isAuthenticated: false,
          isImpersonating: false,
          impersonationToken: null,
          impersonatedUser: null,
          impersonationSessionId: null,
          originalAdminToken: null,
        })
      },

      loadAdmin: async () => {
        const { token, isImpersonating } = get()
        if (!token || isImpersonating) return

        set({ isLoading: true })
        try {
          adminApi.setToken(token)
          const admin = await adminApi.getMe()
          set({ admin, isAuthenticated: true, isLoading: false })
        } catch {
          // Token is invalid, logout
          set({
            admin: null,
            token: null,
            isAuthenticated: false,
            isLoading: false,
          })
        }
      },

      setToken: (token: string) => {
        adminApi.setToken(token)
        set({ token, isAuthenticated: true })
      },

      startImpersonation: async (userId: string, reason: string) => {
        const { token } = get()
        if (!token) throw new Error('Not authenticated')

        try {
          const response: ImpersonationToken = await adminApi.startImpersonation(userId, { reason })

          // Store the original admin token and switch to impersonation token
          set({
            isImpersonating: true,
            impersonationToken: response.token,
            impersonatedUser: response.user,
            impersonationSessionId: response.session_id,
            originalAdminToken: token,
          })

          // Update API client to use impersonation token
          adminApi.setToken(response.token)

          // Return to the main app as the impersonated user
          // The frontend should redirect to the main dashboard
        } catch (error) {
          throw error
        }
      },

      endImpersonation: async () => {
        const { originalAdminToken, impersonationToken } = get()
        if (!originalAdminToken || !impersonationToken) return

        try {
          // End impersonation session using the impersonation token
          adminApi.setToken(impersonationToken)
          await adminApi.endImpersonation()

          // Restore original admin token
          adminApi.setToken(originalAdminToken)
          set({
            isImpersonating: false,
            impersonationToken: null,
            impersonatedUser: null,
            impersonationSessionId: null,
            originalAdminToken: null,
            token: originalAdminToken,
          })
        } catch (error) {
          // Even if the API call fails, restore the admin state
          adminApi.setToken(originalAdminToken)
          set({
            isImpersonating: false,
            impersonationToken: null,
            impersonatedUser: null,
            impersonationSessionId: null,
            originalAdminToken: null,
            token: originalAdminToken,
          })
          throw error
        }
      },
    }),
    {
      name: 'admin-auth-storage',
      partialize: (state) => ({
        token: state.token,
        originalAdminToken: state.originalAdminToken,
        isImpersonating: state.isImpersonating,
        impersonationToken: state.impersonationToken,
        impersonationSessionId: state.impersonationSessionId,
      }),
    }
  )
)
