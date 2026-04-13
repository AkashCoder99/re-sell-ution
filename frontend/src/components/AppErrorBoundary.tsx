import React from 'react'

interface AppErrorBoundaryState {
  hasError: boolean
  message: string
}

export default class AppErrorBoundary extends React.Component<React.PropsWithChildren, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = {
    hasError: false,
    message: ''
  }

  static getDerivedStateFromError(error: Error): AppErrorBoundaryState {
    return {
      hasError: true,
      message: error.message || 'Unexpected runtime error'
    }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo): void {
    console.error('[error] app.render.crash', {
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack
    })
  }

  private handleReload = () => {
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      return (
        <main className="auth-page">
          <section className="auth-card" role="alert" aria-live="assertive">
            <h1>⚠️ ReSellution</h1>
            <p className="error-message">Something went wrong while rendering this page.</p>
            <p className="subtitle">{this.state.message}</p>
            <button type="button" className="auth-form-submit" onClick={this.handleReload}>
              Reload App
            </button>
          </section>
        </main>
      )
    }

    return this.props.children
  }
}
