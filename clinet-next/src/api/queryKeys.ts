export const queryKeys = {
  company: { all: ['company'] as const },
  buildings: { all: ['buildings'] as const },
  inventory: { all: ['inventory'] as const },
  production: { all: ['production'] as const },
  market: {
    all: ['market'] as const,
    ticker: (id: number) => ['market', 'ticker', id] as const,
    depth: (id: number, quality: number) => ['market', 'depth', id, quality] as const,
    orders: (id: number, quality: number) => ['market', 'orders', id, quality] as const,
  },
  research: { all: ['research'] as const, progress: ['research', 'progress'] as const },
  financial: { all: ['financial'] as const },
  executives: {
    all: ['executives'] as const,
    my: ['myExecutives'] as const,
    detail: (id: string) => ['executives', 'detail', id] as const,
  },
  chat: { all: ['chat'] as const },
  leaderboard: { all: ['leaderboard'] as const },
  powerups: { all: ['powerups'] as const },
}
