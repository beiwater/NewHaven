import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/leaderboard/', () => HttpResponse.json({ entries: [], total: 0, page: 1, limit: 10, totalPages: 0, sort: 'net_worth' })),
]
