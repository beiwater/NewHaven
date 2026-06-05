import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v3/resources/', () => {
    return HttpResponse.json({
      resources: [
        { resourceId: 1, name: 'Grain' },
        { resourceId: 2, name: 'Milk' },
        { resourceId: 3, name: 'Flour' },
        { resourceId: 4, name: 'Dough' },
        { resourceId: 5, name: 'Butter' },
        { resourceId: 6, name: 'Sugar' },
        { resourceId: 7, name: 'Cheese' },
        { resourceId: 8, name: 'Steak' },
        { resourceId: 9, name: 'Pizza' },
        { resourceId: 10, name: 'Cake' },
        { resourceId: 11, name: 'Coffee' },
        { resourceId: 12, name: 'Vegetables' },
      ],
    })
  }),

  http.get('/api/v3/market-ticker/:resourceId/', ({ params }) => {
    return HttpResponse.json({
      resource: Number(params.resourceId),
      series: [
        { price: 100, time: new Date(Date.now() - 86400000).toISOString() },
        { price: 102, time: new Date(Date.now() - 43200000).toISOString() },
        { price: 101, time: new Date().toISOString() },
      ],
    })
  }),

  http.get('/api/v3/market-depth/:resourceId/:quality/', () => {
    return HttpResponse.json({
      buys: [{ price: 100, quantity: 10 }],
      sells: [{ price: 110, quantity: 5 }],
    })
  }),

  http.get('/api/v3/market/:resourceId/:quality/', () => {
    return HttpResponse.json([
      { id: 'order-1', resourceId: 1, kind: 'sell', price: 110, quantity: 5, filled: 0, status: 'open' },
    ])
  }),

  http.post('/api/v2/market-order/', () => {
    return HttpResponse.json({ order: { id: 'new-order', status: 'open' } })
  }),

  http.delete('/api/v2/market-order/cancel/:orderId/', () => {
    return HttpResponse.json({ id: 'cancelled', status: 'cancelled' })
  }),

  http.post('/api/v2/market-order/take/', () => {
    return HttpResponse.json({ amountBought: 5, trades: [], moneyDelta: -500 })
  }),
]
