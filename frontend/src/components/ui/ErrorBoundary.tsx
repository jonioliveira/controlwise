'use client'

import { Component, type ReactNode } from 'react'
import { AlertTriangle, RefreshCw } from 'lucide-react'
import { Button } from './Button'

interface ErrorBoundaryProps {
  children: ReactNode
  fallback?: ReactNode
  onReset?: () => void
}

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null })
    this.props.onReset?.()
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div
          role="alert"
          className="flex flex-col items-center justify-center min-h-[200px] p-6 bg-red-50 rounded-lg border border-red-200"
        >
          <AlertTriangle className="h-12 w-12 text-red-500 mb-4" aria-hidden="true" />
          <h2 className="text-lg font-semibold text-red-800 mb-2">
            Algo correu mal
          </h2>
          <p className="text-sm text-red-600 mb-4 text-center max-w-md">
            Ocorreu um erro inesperado. Por favor tente novamente.
          </p>
          <Button
            variant="outline"
            onClick={this.handleReset}
            leftIcon={<RefreshCw className="h-4 w-4" />}
          >
            Tentar novamente
          </Button>
        </div>
      )
    }

    return this.props.children
  }
}

/**
 * Error fallback component for use with error boundaries
 */
interface ErrorFallbackProps {
  error?: Error | null
  resetErrorBoundary?: () => void
  title?: string
  message?: string
}

export function ErrorFallback({
  error,
  resetErrorBoundary,
  title = 'Algo correu mal',
  message = 'Ocorreu um erro inesperado. Por favor tente novamente.',
}: ErrorFallbackProps) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center justify-center min-h-[200px] p-6 bg-red-50 rounded-lg border border-red-200"
    >
      <AlertTriangle className="h-12 w-12 text-red-500 mb-4" aria-hidden="true" />
      <h2 className="text-lg font-semibold text-red-800 mb-2">{title}</h2>
      <p className="text-sm text-red-600 mb-4 text-center max-w-md">{message}</p>
      {error && process.env.NODE_ENV === 'development' && (
        <pre className="text-xs text-red-500 bg-red-100 p-2 rounded mb-4 max-w-full overflow-auto">
          {error.message}
        </pre>
      )}
      {resetErrorBoundary && (
        <Button
          variant="outline"
          onClick={resetErrorBoundary}
          leftIcon={<RefreshCw className="h-4 w-4" />}
        >
          Tentar novamente
        </Button>
      )}
    </div>
  )
}
