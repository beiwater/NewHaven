import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/executives/', () => {
    return HttpResponse.json({ executives: [] })
  }),
  http.get('/api/v2/executives/search/', () => {
    return HttpResponse.json({ candidates: [] })
  }),
]
