import { Component, type ReactNode, type ErrorInfo } from 'react'
import i18n from '@/i18n'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[App Error]', error, info.componentStack)
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback ?? (
          <div className="w-screen h-screen bg-amber-900 flex items-center justify-center text-white">
            <div className="text-center max-w-md">
              <h1 className="text-2xl font-bold mb-2">{i18n.t('error.somethingWrong')}</h1>
              <p className="text-amber-200 text-sm mb-4">
                {this.state.error?.message ?? i18n.t('error.unknown')}
              </p>
              <button
                onClick={() => {
                  this.setState({ hasError: false, error: null })
                  window.location.reload()
                }}
                className="px-4 py-2 bg-amber-600 hover:bg-amber-700 rounded-lg text-sm font-semibold transition-colors"
              >
                {i18n.t('error.reload')}
              </button>
            </div>
          </div>
        )
      )
    }
    return this.props.children
  }
}
