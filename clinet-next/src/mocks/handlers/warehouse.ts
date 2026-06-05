import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/companies/me/warehouse/', () => {
    return HttpResponse.json({
      inventory: [
        { resourceId: 1, quantity: 100 },
        { resourceId: 3, quantity: 50 },
        { resourceId: 8, quantity: 20 },
      ],
      capacity: 500,
      used: 170,
    })
  }),
]
