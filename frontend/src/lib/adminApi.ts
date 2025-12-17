import type {
  SystemAdmin,
  SystemAdminAuthResponse,
  PlatformStats,
  OrganizationWithStats,
  OrganizationModule,
  UserWithOrg,
  AuditLogEntry,
  ImpersonationSessionWithDetails,
  ImpersonationToken,
  RecentActivity,
  AdminLoginRequest,
  AdminChangePasswordRequest,
  AdminCreateOrganizationRequest,
  AdminUpdateOrganizationRequest,
  AdminSuspendRequest,
  AdminStartImpersonationRequest,
  PaginatedResponse,
  ApiResponse,
} from '@/types'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

class AdminApiClient {
  private token: string | null = null

  setToken(token: string | null) {
    this.token = token
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    // Merge with any additional headers from options
    if (options.headers) {
      const optHeaders = options.headers as Record<string, string>
      Object.assign(headers, optHeaders)
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'An error occurred' }))
      throw new Error(error.message || `Request failed with status ${response.status}`)
    }

    return response.json()
  }

  // Auth
  async login(credentials: AdminLoginRequest): Promise<SystemAdminAuthResponse> {
    const response = await this.request<ApiResponse<SystemAdminAuthResponse>>(
      '/admin/auth/login',
      {
        method: 'POST',
        body: JSON.stringify(credentials),
      }
    )
    return response.data
  }

  async getMe(): Promise<SystemAdmin> {
    const response = await this.request<ApiResponse<SystemAdmin>>('/admin/auth/me')
    return response.data
  }

  async changePassword(data: AdminChangePasswordRequest): Promise<void> {
    await this.request('/admin/auth/change-password', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async logout(): Promise<void> {
    await this.request('/admin/auth/logout', { method: 'POST' })
  }

  // Dashboard
  async getStats(): Promise<PlatformStats> {
    const response = await this.request<ApiResponse<PlatformStats>>('/admin/dashboard/stats')
    return response.data
  }

  async getRecentActivity(limit = 10): Promise<RecentActivity[]> {
    const response = await this.request<ApiResponse<RecentActivity[]>>(
      `/admin/dashboard/recent-activity?limit=${limit}`
    )
    return response.data
  }

  // Organizations
  async listOrganizations(params?: {
    page?: number
    limit?: number
    search?: string
    is_active?: boolean
  }): Promise<PaginatedResponse<OrganizationWithStats>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.search) searchParams.set('search', params.search)
    if (params?.is_active !== undefined) searchParams.set('is_active', params.is_active.toString())

    return this.request<PaginatedResponse<OrganizationWithStats>>(
      `/admin/organizations?${searchParams.toString()}`
    )
  }

  async getOrganization(id: string): Promise<OrganizationWithStats> {
    const response = await this.request<ApiResponse<OrganizationWithStats>>(
      `/admin/organizations/${id}`
    )
    return response.data
  }

  async createOrganization(data: AdminCreateOrganizationRequest): Promise<OrganizationWithStats> {
    const response = await this.request<ApiResponse<OrganizationWithStats>>(
      '/admin/organizations',
      {
        method: 'POST',
        body: JSON.stringify(data),
      }
    )
    return response.data
  }

  async updateOrganization(
    id: string,
    data: AdminUpdateOrganizationRequest
  ): Promise<OrganizationWithStats> {
    const response = await this.request<ApiResponse<OrganizationWithStats>>(
      `/admin/organizations/${id}`,
      {
        method: 'PUT',
        body: JSON.stringify(data),
      }
    )
    return response.data
  }

  async suspendOrganization(id: string, data: AdminSuspendRequest): Promise<void> {
    await this.request(`/admin/organizations/${id}/suspend`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async reactivateOrganization(id: string): Promise<void> {
    await this.request(`/admin/organizations/${id}/reactivate`, {
      method: 'POST',
    })
  }

  async deleteOrganization(id: string): Promise<void> {
    await this.request(`/admin/organizations/${id}`, {
      method: 'DELETE',
    })
  }

  // Users
  async listUsers(params?: {
    page?: number
    limit?: number
    search?: string
    organization_id?: string
    is_active?: boolean
    role?: string
  }): Promise<PaginatedResponse<UserWithOrg>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.search) searchParams.set('search', params.search)
    if (params?.organization_id) searchParams.set('organization_id', params.organization_id)
    if (params?.is_active !== undefined) searchParams.set('is_active', params.is_active.toString())
    if (params?.role) searchParams.set('role', params.role)

    return this.request<PaginatedResponse<UserWithOrg>>(
      `/admin/users?${searchParams.toString()}`
    )
  }

  async getUser(id: string): Promise<UserWithOrg> {
    const response = await this.request<ApiResponse<UserWithOrg>>(`/admin/users/${id}`)
    return response.data
  }

  async listOrganizationUsers(
    orgId: string,
    params?: { page?: number; limit?: number }
  ): Promise<PaginatedResponse<UserWithOrg>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())

    return this.request<PaginatedResponse<UserWithOrg>>(
      `/admin/organizations/${orgId}/users?${searchParams.toString()}`
    )
  }

  async suspendUser(id: string, data: AdminSuspendRequest): Promise<void> {
    await this.request(`/admin/users/${id}/suspend`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async reactivateUser(id: string): Promise<void> {
    await this.request(`/admin/users/${id}/reactivate`, {
      method: 'POST',
    })
  }

  async resetUserPassword(id: string, newPassword: string): Promise<void> {
    await this.request(`/admin/users/${id}/reset-password`, {
      method: 'POST',
      body: JSON.stringify({ new_password: newPassword }),
    })
  }

  // Impersonation
  async startImpersonation(
    userId: string,
    data: AdminStartImpersonationRequest
  ): Promise<ImpersonationToken> {
    const response = await this.request<ApiResponse<ImpersonationToken>>(
      `/admin/impersonate/${userId}`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      }
    )
    return response.data
  }

  async endImpersonation(): Promise<void> {
    await this.request('/admin/impersonate/end', {
      method: 'POST',
    })
  }

  async getActiveImpersonationSession(): Promise<ImpersonationSessionWithDetails | null> {
    try {
      const response = await this.request<ApiResponse<ImpersonationSessionWithDetails>>(
        '/admin/impersonate/active'
      )
      return response.data
    } catch {
      return null
    }
  }

  async listImpersonationSessions(params?: {
    page?: number
    limit?: number
    admin_id?: string
  }): Promise<PaginatedResponse<ImpersonationSessionWithDetails>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.admin_id) searchParams.set('admin_id', params.admin_id)

    return this.request<PaginatedResponse<ImpersonationSessionWithDetails>>(
      `/admin/impersonate/sessions?${searchParams.toString()}`
    )
  }

  // Audit Logs
  async listAuditLogs(params?: {
    page?: number
    limit?: number
    admin_id?: string
    action?: string
    entity_type?: string
    entity_id?: string
  }): Promise<PaginatedResponse<AuditLogEntry>> {
    const searchParams = new URLSearchParams()
    if (params?.page) searchParams.set('page', params.page.toString())
    if (params?.limit) searchParams.set('limit', params.limit.toString())
    if (params?.admin_id) searchParams.set('admin_id', params.admin_id)
    if (params?.action) searchParams.set('action', params.action)
    if (params?.entity_type) searchParams.set('entity_type', params.entity_type)
    if (params?.entity_id) searchParams.set('entity_id', params.entity_id)

    return this.request<PaginatedResponse<AuditLogEntry>>(
      `/admin/audit-logs?${searchParams.toString()}`
    )
  }

  // Organization Modules
  async getOrganizationModules(orgId: string): Promise<OrganizationModule[]> {
    const response = await this.request<ApiResponse<OrganizationModule[]>>(
      `/admin/organizations/${orgId}/modules`
    )
    return response.data
  }

  async enableOrganizationModule(orgId: string, moduleName: string): Promise<void> {
    await this.request(`/admin/organizations/${orgId}/modules/${moduleName}/enable`, {
      method: 'POST',
    })
  }

  async disableOrganizationModule(orgId: string, moduleName: string): Promise<void> {
    await this.request(`/admin/organizations/${orgId}/modules/${moduleName}/disable`, {
      method: 'POST',
    })
  }
}

export const adminApi = new AdminApiClient()
