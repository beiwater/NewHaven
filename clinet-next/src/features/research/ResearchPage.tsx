import { useResearch, useStartResearch, useCompleteResearch } from '@/api/hooks/research.hooks'
import { useCompany } from '@/api/hooks/company.hooks'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

export function ResearchPage() {
  const { data: projects, isLoading } = useResearch()
  const startResearch = useStartResearch()
  const completeResearch = useCompleteResearch()
  const { data: company } = useCompany()
  const companyData = company as unknown as { authCompany?: { money?: number } } | undefined
  const money = companyData?.authCompany?.money ?? 0

  if (isLoading) return <div className="p-4 text-xs text-amber-500">Loading research...</div>

  const allProjects = projects ?? []
  const available = allProjects.filter((p) => p.status === 'available' || p.status === 'locked')
  const inProgress = allProjects.filter((p) => p.status === 'in_progress')

  return (
    <div className="p-4 space-y-4">
      <h2 className="text-lg font-bold text-amber-900">Research</h2>

      {inProgress.length > 0 && (
        <div>
          <h3 className="text-xs font-semibold text-blue-700 uppercase mb-2">In Progress</h3>
          {inProgress.map((p) => (
            <Card key={p.id} className="mb-2">
              <CardContent className="p-3">
                <div className="flex justify-between items-center">
                  <div>
                    <div className="text-sm font-semibold text-amber-900">{p.name}</div>
                    <div className="text-[10px] text-amber-600">{Math.round(p.progress)}%</div>
                  </div>
                  <Button size="sm" onClick={() => completeResearch.mutate(p.id)}>Complete Now</Button>
                </div>
                <div className="mt-2 h-1.5 bg-amber-200 rounded-full overflow-hidden">
                  <div className="h-full bg-blue-500 rounded-full" style={{ width: `${p.progress}%` }} />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <div>
        <h3 className="text-xs font-semibold text-amber-700 uppercase mb-2">Available Projects</h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
          {available.map((p) => {
            const canAfford = (p.cashCost ?? 0) <= money
            const locked = p.status === 'locked'
            return (
              <Card key={p.id} className={locked ? 'opacity-50' : ''}>
                <CardContent className="p-3">
                  <div className="text-sm font-semibold text-amber-900">{p.name}</div>
                  {p.cashCost && <div className="text-[10px] text-amber-600">Cost: ${p.cashCost.toLocaleString()}</div>}
                  {p.durationHours && <div className="text-[10px] text-amber-600">Duration: {p.durationHours}h</div>}
                  <Button
                    size="sm"
                    className="mt-2"
                    disabled={locked || !canAfford || startResearch.isPending}
                    onClick={() => startResearch.mutate(p.id)}
                  >
                    {locked ? 'Locked' : canAfford ? 'Start' : 'Need More Money'}
                  </Button>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </div>
    </div>
  )
}
