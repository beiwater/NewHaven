import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/research/', () => HttpResponse.json({ projects: [] })),
  http.get('/api/v2/research/progress/', () => HttpResponse.json({ projects: [] })),
  http.post('/api/v2/research/start/', () => HttpResponse.json({ project: { id: 'r1', status: 'in_progress' }, status: 'ok' })),
  http.post('/api/v2/research/complete/:id/', () => HttpResponse.json({ ok: true, projectId: 'r1', patentsGained: 1, qualityImproved: 1 })),
]
