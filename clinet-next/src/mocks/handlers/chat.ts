import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/chat/', () => {
    return HttpResponse.json({ messages: [] })
  }),
]
