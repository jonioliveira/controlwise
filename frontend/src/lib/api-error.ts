import { AxiosError } from 'axios'

export interface ApiErrorResponse {
  error: string
  message: string
  details?: Record<string, string[]>
}

export class ApiError extends Error {
  public readonly status: number
  public readonly code: string
  public readonly details?: Record<string, string[]>

  constructor(message: string, status: number, code: string, details?: Record<string, string[]>) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }

  static fromAxiosError(error: AxiosError<ApiErrorResponse>): ApiError {
    const status = error.response?.status || 500
    const data = error.response?.data

    // Handle different error scenarios
    if (!error.response) {
      // Network error
      return new ApiError(
        'Não foi possível conectar ao servidor. Verifique sua conexão.',
        0,
        'NETWORK_ERROR'
      )
    }

    if (status === 401) {
      return new ApiError(
        'Sessão expirada. Por favor faça login novamente.',
        401,
        'UNAUTHORIZED'
      )
    }

    if (status === 403) {
      return new ApiError(
        data?.message || 'Não tem permissão para realizar esta ação.',
        403,
        'FORBIDDEN'
      )
    }

    if (status === 404) {
      return new ApiError(
        data?.message || 'Recurso não encontrado.',
        404,
        'NOT_FOUND'
      )
    }

    if (status === 409) {
      return new ApiError(
        data?.message || 'Conflito ao processar o pedido.',
        409,
        'CONFLICT',
        data?.details
      )
    }

    if (status === 422) {
      return new ApiError(
        data?.message || 'Dados inválidos.',
        422,
        'VALIDATION_ERROR',
        data?.details
      )
    }

    if (status >= 500) {
      return new ApiError(
        'Ocorreu um erro no servidor. Por favor tente mais tarde.',
        status,
        'SERVER_ERROR'
      )
    }

    return new ApiError(
      data?.message || 'Ocorreu um erro inesperado.',
      status,
      data?.error || 'UNKNOWN_ERROR',
      data?.details
    )
  }

  /**
   * Get a user-friendly error message
   */
  getUserMessage(): string {
    return this.message
  }

  /**
   * Check if this is a validation error
   */
  isValidationError(): boolean {
    return this.status === 422
  }

  /**
   * Get field-specific errors for form validation
   */
  getFieldErrors(): Record<string, string> {
    if (!this.details) return {}

    const fieldErrors: Record<string, string> = {}
    for (const [field, messages] of Object.entries(this.details)) {
      fieldErrors[field] = messages[0] // Get first error message for each field
    }
    return fieldErrors
  }
}

/**
 * Helper to check if an error is an ApiError
 */
export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

/**
 * Extract user-friendly message from any error
 */
export function getErrorMessage(error: unknown): string {
  if (isApiError(error)) {
    return error.getUserMessage()
  }

  if (error instanceof Error) {
    return error.message
  }

  return 'Ocorreu um erro inesperado.'
}
