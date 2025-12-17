import axios, { AxiosInstance, AxiosError } from 'axios'
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  User,
  Client,
  WorkSheet,
  Budget,
  Project,
  Task,
  Payment,
  Notification,
  ApiResponse,
} from '@/types'
import { useAuthStore } from '@/stores/authStore'
import { ApiError, type ApiErrorResponse } from './api-error'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 30000, // 30 seconds
    })

    // Request interceptor to add auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = this.getToken()
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError<ApiErrorResponse>) => {
        if (error.response?.status === 401) {
          // Clear auth state on unauthorized
          useAuthStore.getState().logout()
          if (typeof window !== 'undefined' && !window.location.pathname.includes('/login')) {
            window.location.href = '/login'
          }
        }
        // Convert to ApiError for consistent error handling
        throw ApiError.fromAxiosError(error)
      }
    )
  }

  private getToken(): string | null {
    return useAuthStore.getState().token
  }

  // Auth endpoints
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.client.post<ApiResponse<AuthResponse>>('/auth/login', data)
    return response.data.data
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.client.post<ApiResponse<AuthResponse>>('/auth/register', data)
    return response.data.data
  }

  async logout(): Promise<void> {
    try {
      await this.client.post('/auth/logout')
    } catch {
      // Ignore logout errors
    }
  }

  async getCurrentUser(): Promise<User> {
    const response = await this.client.get<ApiResponse<User>>('/auth/me')
    return response.data.data
  }

  // Clients
  async getClients(): Promise<Client[]> {
    const response = await this.client.get<ApiResponse<Client[]>>('/clients')
    return response.data.data
  }

  async createClient(data: Partial<Client>): Promise<Client> {
    const response = await this.client.post<ApiResponse<Client>>('/clients', data)
    return response.data.data
  }

  async updateClient(id: string, data: Partial<Client>): Promise<Client> {
    const response = await this.client.put<ApiResponse<Client>>(`/clients/${id}`, data)
    return response.data.data
  }

  async deleteClient(id: string): Promise<void> {
    await this.client.delete(`/clients/${id}`)
  }

  // Worksheets
  async getWorksheets(): Promise<WorkSheet[]> {
    const response = await this.client.get<ApiResponse<WorkSheet[]>>('/worksheets')
    return response.data.data
  }

  async createWorksheet(data: Partial<WorkSheet>): Promise<WorkSheet> {
    const response = await this.client.post<ApiResponse<WorkSheet>>('/worksheets', data)
    return response.data.data
  }

  async updateWorksheet(id: string, data: Partial<WorkSheet>): Promise<WorkSheet> {
    const response = await this.client.put<ApiResponse<WorkSheet>>(`/worksheets/${id}`, data)
    return response.data.data
  }

  async deleteWorksheet(id: string): Promise<void> {
    await this.client.delete(`/worksheets/${id}`)
  }

  // Budgets
  async getBudgets(): Promise<Budget[]> {
    const response = await this.client.get<ApiResponse<Budget[]>>('/budgets')
    return response.data.data
  }

  async createBudget(data: Partial<Budget>): Promise<Budget> {
    const response = await this.client.post<ApiResponse<Budget>>('/budgets', data)
    return response.data.data
  }

  async updateBudget(id: string, data: Partial<Budget>): Promise<Budget> {
    const response = await this.client.put<ApiResponse<Budget>>(`/budgets/${id}`, data)
    return response.data.data
  }

  async approveBudget(id: string): Promise<void> {
    await this.client.post(`/budgets/${id}/approve`)
  }

  async rejectBudget(id: string, notes: string): Promise<void> {
    await this.client.post(`/budgets/${id}/reject`, { rejection_notes: notes })
  }

  // Projects
  async getProjects(): Promise<Project[]> {
    const response = await this.client.get<ApiResponse<Project[]>>('/projects')
    return response.data.data
  }

  async createProject(data: Partial<Project>): Promise<Project> {
    const response = await this.client.post<ApiResponse<Project>>('/projects', data)
    return response.data.data
  }

  async updateProject(id: string, data: Partial<Project>): Promise<Project> {
    const response = await this.client.put<ApiResponse<Project>>(`/projects/${id}`, data)
    return response.data.data
  }

  async updateProjectProgress(id: string, progress: number): Promise<void> {
    await this.client.patch(`/projects/${id}/progress`, { progress })
  }

  // Tasks
  async getTasks(projectId?: string): Promise<Task[]> {
    const params = projectId ? { project_id: projectId } : {}
    const response = await this.client.get<ApiResponse<Task[]>>('/tasks', { params })
    return response.data.data
  }

  async createTask(data: Partial<Task>): Promise<Task> {
    const response = await this.client.post<ApiResponse<Task>>('/tasks', data)
    return response.data.data
  }

  async updateTask(id: string, data: Partial<Task>): Promise<Task> {
    const response = await this.client.put<ApiResponse<Task>>(`/tasks/${id}`, data)
    return response.data.data
  }

  async deleteTask(id: string): Promise<void> {
    await this.client.delete(`/tasks/${id}`)
  }

  // Payments
  async getPayments(): Promise<Payment[]> {
    const response = await this.client.get<ApiResponse<Payment[]>>('/payments')
    return response.data.data
  }

  async createPayment(data: Partial<Payment>): Promise<Payment> {
    const response = await this.client.post<ApiResponse<Payment>>('/payments', data)
    return response.data.data
  }

  async markPaymentAsPaid(id: string): Promise<void> {
    await this.client.post(`/payments/${id}/mark-paid`)
  }

  // Notifications
  async getNotifications(): Promise<Notification[]> {
    const response = await this.client.get<ApiResponse<Notification[]>>('/notifications')
    return response.data.data
  }

  async markNotificationAsRead(id: string): Promise<void> {
    await this.client.post(`/notifications/${id}/read`)
  }

  async markAllNotificationsAsRead(): Promise<void> {
    await this.client.post('/notifications/read-all')
  }

  async getUnreadCount(): Promise<number> {
    const response = await this.client.get<ApiResponse<{ count: number }>>('/notifications/unread-count')
    return response.data.data.count
  }

  // Generic HTTP methods for flexible API calls
  async get<T = unknown>(url: string, config?: Parameters<typeof this.client.get>[1]) {
    return this.client.get<T>(url, config)
  }

  async post<T = unknown>(url: string, data?: unknown, config?: Parameters<typeof this.client.post>[2]) {
    return this.client.post<T>(url, data, config)
  }

  async put<T = unknown>(url: string, data?: unknown, config?: Parameters<typeof this.client.put>[2]) {
    return this.client.put<T>(url, data, config)
  }

  async patch<T = unknown>(url: string, data?: unknown, config?: Parameters<typeof this.client.patch>[2]) {
    return this.client.patch<T>(url, data, config)
  }

  async delete<T = unknown>(url: string, config?: Parameters<typeof this.client.delete>[1]) {
    return this.client.delete<T>(url, config)
  }
}

export const api = new ApiClient()
