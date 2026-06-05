import { http, HttpResponse } from 'msw'

export const handlers = [
  http.post('/api/v2/executives/search/', () => HttpResponse.json({ candidates: [] })),
  http.post('/api/v2/executives/recruit/', () => HttpResponse.json({ executive: { id: 'e1' }, cost: 5000 })),
  http.get('/api/v2/executives/', () => HttpResponse.json([])),
  http.post('/api/v2/executives/train/:id/', () => HttpResponse.json({ executiveId: 'e1', newLevel: 2, cost: 5000, timeSeconds: 3600 })),
  http.get('/api/v3/executives/:id/', () => HttpResponse.json({ id: 'e1', name: 'Executive', rarity: 'Rare', level: 1, skills: { production: 5, sales: 3, management: 2 }, salary: 600, stats: [] })),
]
