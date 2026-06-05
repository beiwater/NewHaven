import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/production/jobs/', () => {
    return HttpResponse.json([
      {
        id: 'job-1',
        buildingId: 'b-1',
        resourceId: 1,
        amount: 100,
        claimedAmount: 60,
        claimableAmount: 0,
        status: 'in_progress',
        completesAt: new Date(Date.now() + 3600000).toISOString(),
      },
    ])
  }),

  http.get('/api/v2/production/queue/', () => {
    return HttpResponse.json({
      byBuilding: { 'b-1': { inUse: 1, maxSlots: 3 } },
      inUse: 1,
      maxSlots: 3,
    })
  }),

  http.get('/api/v2/production/claimable/', () => {
    return HttpResponse.json([
      {
        id: 'job-claim-1',
        buildingId: 'b-1',
        resourceId: 8,
        amount: 20,
        claimedAmount: 20,
        claimableAmount: 20,
        status: 'completed',
        completesAt: new Date().toISOString(),
      },
    ])
  }),

  http.get('/api/v2/buildings/:buildingId/production-options/', () => {
    return HttpResponse.json([
      { resourceId: 1, name: 'Grain', producedPerHour: 60, sourcingCost: 10 },
      { resourceId: 2, name: 'Milk', producedPerHour: 40, sourcingCost: 15 },
    ])
  }),

  http.post('/api/v1/buildings/:buildingId/busy/', () => {
    return HttpResponse.json({ status: 'started' })
  }),

  http.post('/api/v2/production/claim/:jobId/', () => {
    return HttpResponse.json({ jobId: 'claim-1', status: 'claimed', output: { 8: 20 }, quality: 1 })
  }),

  http.post('/api/v2/production/claim-all/', () => {
    return HttpResponse.json({ claimed: [], errors: [], total: 0 })
  }),

  http.post('/api/v2/production/cancel/', () => {
    return HttpResponse.json({ jobId: 'cancelled', status: 'cancelled' })
  }),
]
