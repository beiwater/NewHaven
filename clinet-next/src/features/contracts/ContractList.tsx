import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

interface DailyOrder {
  id: string
  resourceId: number
  quality: number
  quantity: number
  cashReward?: number
  rewardCash?: number
  rewardXP: number
  status: string
}

interface GovContract {
  id: string
  resourceId: number
  quality: number
  quantity: number
  name?: string
  value?: number
  status: string
}

export function ContractList() {
  const qc = useQueryClient()
  const { data: dailyData } = useQuery({
    queryKey: queryKeys.contracts.daily(),
    queryFn: () => api.get<{ orders: DailyOrder[]; date: string }>('/api/v2/orders/daily/'),
  })
  const { data: govData } = useQuery({
    queryKey: queryKeys.contracts.gov(),
    queryFn: () => api.get<GovContract[]>('/api/v3/government-orders/'),
  })

  const completeOrder = useMutation({
    mutationFn: (id: string) => api.post('/api/v2/orders/daily/complete/', { orderId: id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.contracts.daily() }),
  })

  const orders = dailyData?.orders ?? []
  const contracts = govData ?? []

  return (
    <div className="p-4 space-y-4">
      <h2 className="text-lg font-bold text-amber-900">Contracts</h2>

      {orders.length > 0 && (
        <div>
          <h3 className="text-xs font-semibold text-amber-700 uppercase mb-2">Daily Orders</h3>
          <div className="space-y-2">
            {orders.map((o) => (
              <Card key={o.id}>
                <CardContent className="p-3 flex justify-between items-center">
                  <div>
                    <div className="text-sm font-semibold text-amber-900">Order #{o.id?.slice(0, 8)}</div>
                    <div className="text-[10px] text-amber-600">Reward: ${o.cashReward ?? 0}</div>
                  </div>
                  <Button size="sm" onClick={() => completeOrder.mutate(o.id)}>Complete</Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {contracts.length > 0 && (
        <div>
          <h3 className="text-xs font-semibold text-amber-700 uppercase mb-2">Government Contracts</h3>
          <div className="space-y-2">
            {contracts.map((c) => (
              <Card key={c.id}>
                <CardContent className="p-3">
                  <div className="text-sm font-semibold text-amber-900">{c.name ?? `Contract ${c.id}`}</div>
                  <div className="text-[10px] text-amber-600">Value: ${c.value?.toLocaleString() ?? 0}</div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {orders.length === 0 && contracts.length === 0 && (
        <div className="text-xs text-amber-400 italic">No contracts available</div>
      )}
    </div>
  )
}
