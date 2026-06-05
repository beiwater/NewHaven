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
}
