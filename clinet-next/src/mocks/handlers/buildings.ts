import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/companies/me/buildings/', () => {
    return HttpResponse.json([
      { id: 'b-1', kind: 1, name: 'Farm', level: 1, placed: true, x: 0, y: 0, mapId: 'harbor', slotId: 's1' },
      { id: 'b-2', kind: 3, name: 'Mill', level: 2, placed: true, x: 1, y: 0, mapId: 'harbor', slotId: 's2' },
    ])
  }),

  http.get('/api/v2/companies/:companyId/buildings/', () => {
    return HttpResponse.json([
      { id: 'b-1', kind: 1, name: 'Farm', level: 1, placed: true },
    ])
  }),

  http.post('/api/v2/buildings/buy/', () => {
    return HttpResponse.json({
      building: { id: 'b-new', kind: 2, name: 'Barn', level: 1, placed: false },
      cost: 1000,
      money: 9000,
    })
  }),

  http.post('/api/v2/buildings/place/', () => {
    return HttpResponse.json({ building: { id: 'b-new', placed: true }, money: 9000 })
  }),

  http.post('/api/v2/buildings/move/', () => {
    return HttpResponse.json({ building: { id: 'b-moved' } })
  }),

  http.post('/api/v1/buildings/:buildingId/upgrade/', () => {
    return HttpResponse.json({ buildingId: 'b-1', oldLevel: 1, newLevel: 2, cost: 2000, outputMultiplier: 1.5 })
  }),

  http.post('/api/v2/buildings/demolish/', () => {
    return HttpResponse.json({ refund: 500, status: 'demolished' })
  }),
]
