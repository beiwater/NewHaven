import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/messages/', () => HttpResponse.json([])),
  http.post('/api/v2/message/', () => HttpResponse.json({ id: 'm1', status: 'sent' })),
  http.get('/api/v2/message/:id/read/', () => HttpResponse.json({ ok: true })),
  http.get('/api/v2/chatroom/', () => HttpResponse.json([])),
  http.get('/api/v2/contacts/', () => HttpResponse.json({ contacts: [] })),
]
