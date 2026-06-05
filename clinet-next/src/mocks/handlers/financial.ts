import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v2/financial/income/', () => HttpResponse.json({ revenue: 0, expenses: 0, net: 0 })),
  http.get('/api/v2/financial/balance/', () => HttpResponse.json({ cash: 10000, assets: 50000, liabilities: 0 })),
  http.get('/api/v2/financial/cashflow/', () => HttpResponse.json({ inflows: [], outflows: [], net: 0 })),
  http.get('/api/v2/financial/recent/', () => HttpResponse.json({ entries: [] })),
  http.get('/api/v2/financial/history/', () => HttpResponse.json({ series: [] })),
  http.get('/api/v2/financial/overview/', () => HttpResponse.json({ cash: 10000, netWorth: 50000, revenue: 0, expenses: 0 })),
]
