import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/research/', () => {
    return HttpResponse.json({ projects: [] })
  }),
]
