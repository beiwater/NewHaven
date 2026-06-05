export const queryKeys = {
  company: {
    all: () => ['company'] as const,
  },
  buildings: {
    all: () => ['buildings'] as const,
    byCompany: (companyId: number | null) => ['buildings', 'company', companyId] as const,
    productionOptions: (buildingId: string | undefined) => ['buildings', 'production-options', buildingId] as const,
  },
  inventory: {
    warehouse: () => ['warehouse'] as const,
  },
  production: {
    jobs: () => ['production', 'jobs'] as const,
    queue: () => ['production', 'queue'] as const,
    claimable: () => ['production', 'claimable'] as const,
  },
  market: {
    resources: () => ['market', 'resources'] as const,
    ticker: (resourceId: number) => ['market', 'ticker', resourceId] as const,
    depth: (resourceId: number, quality: number) => ['market', 'depth', resourceId, quality] as const,
    orders: (resourceId: number, quality: number) => ['market', 'orders', resourceId, quality] as const,
  },
  research: {
    all: () => ['research'] as const,
    progress: () => ['research', 'progress'] as const,
  },
  financial: {
    income: () => ['financial', 'income'] as const,
    balance: () => ['financial', 'balance'] as const,
    cashflow: () => ['financial', 'cashflow'] as const,
    recentCashflow: () => ['financial', 'recent-cashflow'] as const,
    pastFinances: () => ['financial', 'past'] as const,
    overview: () => ['financial', 'overview'] as const,
  },
  executives: {
    all: () => ['executives'] as const,
    my: () => ['myExecutives'] as const,
    detail: (id: string | null) => ['executives', 'detail', id] as const,
  },
  chat: {
    messages: () => ['chat', 'messages'] as const,
    chatroom: () => ['chat', 'chatroom'] as const,
    contacts: () => ['chat', 'contacts'] as const,
  },
  powerups: {
    types: () => ['powerupTypes'] as const,
    active: () => ['activePowerup'] as const,
  },
  leaderboard: {
    all: (sort: string, page: number, limit: number) => ['leaderboard', sort, page, limit] as const,
  },
  contracts: {
    daily: () => ['dailyOrders'] as const,
    gov: () => ['govContracts'] as const,
  },
}
