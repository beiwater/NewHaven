import { usePowerupTypes, useActivePowerup, useActivatePowerup } from '@/api/hooks/powerup.hooks'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export function PowerPanel() {
  const { data: typesData } = usePowerupTypes()
  const { data: activeData } = useActivePowerup()
  const activate = useActivatePowerup()

  const types = typesData?.boosts ?? []
  const active = activeData?.active ?? []
  const remaining = activeData?.remaining ?? 0

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-bold text-amber-900">Power-ups</h2>
        <Badge variant="outline">{remaining} uses left</Badge>
      </div>

      {active.length > 0 && (
        <div>
          <h3 className="text-xs font-semibold text-green-700 uppercase mb-2">Active</h3>
          {active.map((a, i) => (
            <Card key={i} className="mb-2 border-green-200">
              <CardContent className="p-3 flex justify-between items-center">
                <span className="text-sm font-semibold text-amber-900">{a.type}</span>
                <span className="text-xs text-amber-600">Ends {new Date(a.endsAt).toLocaleTimeString()}</span>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {types.map((t) => (
          <Card key={t.id}>
            <CardContent className="p-3">
              <div className="text-sm font-semibold text-amber-900">{t.name}</div>
              <div className="text-[10px] text-amber-600 mb-2">{t.desc}</div>
              <Button size="sm" onClick={() => activate.mutate(t.id)} disabled={activate.isPending || remaining <= 0}>
                {remaining <= 0 ? 'No Uses' : 'Activate'}
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
