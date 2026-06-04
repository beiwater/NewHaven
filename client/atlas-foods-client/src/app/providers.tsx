import { useEffect, type ReactNode } from 'react'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import '@/i18n'
import { ApiError } from '@/api/client'
import { AudioProvider } from '@/audio/useAudio'
import { audio } from '@/audio/AudioManager'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status === 401) return false
        return failureCount < 1
      },
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status === 401) return false
        return failureCount < 1
      },
    },
  },
})

export function Providers({ children }: { children: ReactNode }) {
  useEffect(() => {
    audio.init()
  }, [])

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AudioProvider>{children}</AudioProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
