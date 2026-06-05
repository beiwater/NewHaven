import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/players/simboosts/', () => HttpResponse.json({ boosts: [] })),
  http.get('/api/v2/players/simboosts-use/', () => HttpResponse.json({ remaining: 3, active: [] })),
  http.post('/api/v2/players/simboosts-use/', () => HttpResponse.json({ boostId: 'speed', endsAt: new Date(Date.now() + 3600000).toISOString(), multiplier: 2 })),
]
