import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v3/market-ticker/:resourceId/', ({ params }) => {
    return HttpResponse.json({
      resource: Number(params.resourceId),
      series: [{ price: 100, time: new Date().toISOString() }],
    })
  }),
]
