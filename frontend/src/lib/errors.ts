import { AxiosError } from 'axios'

export interface ApiError {
  error: string
  code: string
  message: string
  details?: Array<{ field: string; message: string }>
}

export function getErrorMessage(error: unknown): string {
  if (error instanceof AxiosError) {
    const apiError = error.response?.data as ApiError | undefined

    // Handle validation errors
    if (apiError?.details && apiError.details.length > 0) {
      return apiError.details.map((d) => d.message).join(', ')
    }

    // Handle API error message
    if (apiError?.message) {
      return apiError.message
    }

    // Handle network errors
    if (error.code === 'ERR_NETWORK') {
      return 'Sem conexão à internet. Verifique a sua ligação.'
    }

    // Handle timeout
    if (error.code === 'ECONNABORTED') {
      return 'A requisição demorou demasiado. Tente novamente.'
    }

    // Handle specific HTTP status codes
    switch (error.response?.status) {
      case 400:
        return 'Dados inválidos. Verifique os campos preenchidos.'
      case 401:
        return 'Sessão expirada. Por favor, faça login novamente.'
      case 403:
        return 'Não tem permissões para realizar esta ação.'
      case 404:
        return 'Recurso não encontrado.'
      case 409:
        return 'Conflito de dados. O recurso já existe.'
      case 429:
        return 'Muitas requisições. Aguarde um momento.'
      case 500:
        return 'Erro interno do servidor. Tente novamente mais tarde.'
      default:
        return 'Ocorreu um erro. Tente novamente.'
    }
  }

  if (error instanceof Error) {
    return error.message
  }

  return 'Ocorreu um erro desconhecido.'
}

export function getFieldErrors(error: unknown): Record<string, string> {
  const fieldErrors: Record<string, string> = {}

  if (error instanceof AxiosError) {
    const apiError = error.response?.data as ApiError | undefined

    if (apiError?.details) {
      for (const detail of apiError.details) {
        fieldErrors[detail.field] = detail.message
      }
    }
  }

  return fieldErrors
}

export function isNetworkError(error: unknown): boolean {
  return error instanceof AxiosError && error.code === 'ERR_NETWORK'
}

export function isUnauthorizedError(error: unknown): boolean {
  return error instanceof AxiosError && error.response?.status === 401
}

export function isValidationError(error: unknown): boolean {
  if (error instanceof AxiosError) {
    const apiError = error.response?.data as ApiError | undefined
    return apiError?.code === 'VALIDATION_ERROR'
  }
  return false
}
