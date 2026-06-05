import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/financial/', () => {
    return HttpResponse.json({ transactions: [], summary: { cash: 0, revenue: 0, expenses: 0 } })
  }),
]
