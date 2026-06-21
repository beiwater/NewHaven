import React from 'react'
import ReactDOM from 'react-dom/client'
import { Providers } from './app/providers'
import { App } from './app/App'
import './styles/globals.css'

async function bootstrap() {
  if (import.meta.env.DEV) {
    // Auto-login in dev mode (same as original AuthGate)
    const token = localStorage.getItem('atlas_auth_token')
    if (!token) {
      try {
        const res = await fetch('/api/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username: 'dev', password: 'dev' }),
        })
        if (res.ok) {
          const data = await res.json()
          localStorage.setItem('atlas_auth_token', data.token)
          localStorage.setItem('atlas_company_id', String(data.companyId))
        }
      } catch { /* offline — rely on MSW or fallback */ }
    }

    if (import.meta.env.VITE_ENABLE_MSW === 'true') {
      const { worker } = await import('./mocks/browser')
      await worker.start({ onUnhandledRequest: 'bypass' })
    }
  }

  ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
      <Providers>
        <App />
      </Providers>
    </React.StrictMode>,
  )
}

bootstrap()
