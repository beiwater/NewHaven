import { useMutation } from '@tanstack/react-query'
import { api } from './client'

export type ReportCategory = 'bug' | 'feature' | 'feedback' | 'other'

export interface ReportPayload {
  category: ReportCategory
  description: string
}

export interface SubmitReportResponse {
  id: string
  status: string
}

const CATEGORY_LABELS: Record<ReportCategory, string> = {
  bug: 'Bug Report',
  feature: 'Feature Request',
  feedback: 'Feedback',
  other: 'Other',
}

export { CATEGORY_LABELS }

export function useSubmitReport() {
  return useMutation<SubmitReportResponse, Error, ReportPayload>({
    mutationFn: (payload) => api.post<SubmitReportResponse>('/api/v2/report/', payload),
  })
}
