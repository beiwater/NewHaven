import { useState, useMemo } from 'react'
import {
  useResearch,
  useResearchProgress,
  useStartResearch,
  useCompleteResearch,
} from '@/api/research.api'
import type { ResearchProject } from '@/api/research.api'
import { FlaskIcon, GearIcon, PriceTagIcon, TruckIcon, CutleryIcon, BankIcon, CheckIcon, LockIcon, ClockIcon } from './icons'

// =========== Helpers ===========

function fmtCash(n: number | undefined): string {
  if (n === undefined) return ''
  return `$${n.toLocaleString()}`
}

function fmtDuration(hours: number | undefined): string {
  if (hours === undefined || hours <= 0) return ''
  const h = Math.floor(hours)
  const m = Math.round((hours - h) * 60)
  if (h > 0 && m > 0) return `${h}h ${m}m`
  if (h > 0) return `${h}h`
  return `${m}m`
}

function timeRemaining(completesAt: string | undefined): string {
  if (!completesAt) return ''
  const diff = new Date(completesAt).getTime() - Date.now()
  if (diff <= 0) return 'Completing...'
  const h = Math.floor(diff / 3_600_000)
  const m = Math.floor((diff % 3_600_000) / 60_000)
  const s = Math.floor((diff % 60_000) / 1000)
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

// =========== Category config ===========

const CATEGORIES = [
  { id: 'production', label: 'Production', icon: <GearIcon /> },
  { id: 'sales', label: 'Sales', icon: <PriceTagIcon /> },
  { id: 'logistics', label: 'Logistics', icon: <TruckIcon /> },
  { id: 'restaurant', label: 'Restaurant', icon: <CutleryIcon /> },
  { id: 'finance', label: 'Finance', icon: <BankIcon /> },
] as const

type CategoryId = (typeof CATEGORIES)[number]['id']

// Map research names to categories (backend doesn't return category, so we infer)
function inferCategory(name: string): CategoryId {
  const lower = name.toLowerCase()
  if (lower.includes('grain') || lower.includes('plant') || lower.includes('refrigeration') || lower.includes('packaging')) return 'production'
  if (lower.includes('sale') || lower.includes('price') || lower.includes('premium')) return 'sales'
  if (lower.includes('routing') || lower.includes('logistics') || lower.includes('delivery') || lower.includes('smart')) return 'logistics'
  if (lower.includes('staff') || lower.includes('incentive') || lower.includes('restaurant') || lower.includes('employee')) return 'restaurant'
  if (lower.includes('bookkeeping') || lower.includes('finance') || lower.includes('automated') || lower.includes('accounting')) return 'finance'
  if (lower.includes('chemical') || lower.includes('mining') || lower.includes('energy')) return 'production'
  return 'production'
}

// =========== Card component ===========

function ResearchCard({
  project,
  isSelected,
  onSelect,
  onStart,
  onComplete,
  pendingAction,
}: {
  project: ResearchProject
  isSelected: boolean
  onSelect: () => void
  onStart: () => void
  onComplete: () => void
  pendingAction: boolean
}) {
  const { name, status, progress, cashCost, durationHours, resourceCost } = project

  const isAvailable = status === 'available'
  const isInProgress = status === 'in_progress'
  const isCompleted = status === 'completed'
  const isLocked = status === 'locked'

  const borderColor = isSelected
    ? 'border-blue-500/60 ring-2 ring-blue-300/40'
    : isCompleted
      ? 'border-green-400/50'
      : isInProgress
        ? 'border-blue-400/50'
        : 'border-amber-300/40'

  return (
    <button
      onClick={onSelect}
      className={`relative flex flex-col rounded-lg border-2 bg-amber-50/70 p-3 text-left transition-all duration-150 hover:shadow-md ${borderColor} ${isCompleted ? 'opacity-80' : ''}`}
    >
      {/* Status badge */}
      <div className="flex items-center justify-between mb-1.5">
        <span className="inline-flex items-center gap-1 rounded bg-amber-200/60 px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wider text-amber-700">
          {inferCategory(name)}
        </span>
        {isCompleted && (
          <span className="inline-flex items-center gap-0.5 text-[10px] font-bold text-green-600">
            <CheckIcon className="w-3 h-3" /> Done
          </span>
        )}
        {isInProgress && (
          <span className="text-[10px] font-bold text-blue-600">{progress}%</span>
        )}
      </div>

      {/* Name */}
      <h3 className="text-sm font-bold text-amber-900 leading-tight mb-1.5">{name}</h3>

      {/* Cost row (if available) */}
      <div className="space-y-0.5 mb-2">
        {cashCost !== undefined && cashCost > 0 && (
          <div className="flex items-center gap-1 text-[10px] text-amber-700">
            <span className="font-semibold text-green-700">{fmtCash(cashCost)}</span>
            {durationHours !== undefined && durationHours > 0 && (
              <>
                <span className="text-amber-400">·</span>
                <ClockIcon />
                <span>{fmtDuration(durationHours)}</span>
              </>
            )}
          </div>
        )}
        {resourceCost && Object.keys(resourceCost).length > 0 && (
          <div className="flex flex-wrap gap-1">
            {Object.entries(resourceCost).map(([rid, qty]) => (
              <span key={rid} className="text-[9px] bg-amber-100/70 rounded px-1 py-0.5 text-amber-600">
                {rid}: {qty}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Action / Status area */}
      {isAvailable && (
        <button
          onClick={(e) => { e.stopPropagation(); onStart() }}
          disabled={pendingAction}
          className="mt-auto w-full rounded bg-green-600 py-1.5 text-[11px] font-bold text-white hover:bg-green-700 disabled:opacity-50 transition-colors"
        >
          {pendingAction ? 'Starting...' : 'Start Research'}
        </button>
      )}

      {isInProgress && (
        <div className="mt-auto space-y-1">
          <div className="h-1.5 w-full rounded-full bg-amber-200/60 overflow-hidden">
            <div
              className="h-full rounded-full bg-blue-500 transition-all duration-500"
              style={{ width: `${Math.min(progress, 100)}%` }}
            />
          </div>
          <div className="flex items-center justify-between text-[9px] text-amber-500">
            <span>Researching...</span>
            <span>{timeRemaining(project.completesAt)}</span>
          </div>
          <button
            onClick={(e) => { e.stopPropagation(); onComplete() }}
            disabled={pendingAction}
            className="w-full rounded bg-amber-500 py-1 text-[10px] font-bold text-white hover:bg-amber-600 disabled:opacity-50 transition-colors"
          >
            {pendingAction ? 'Completing...' : 'Complete Now'}
          </button>
        </div>
      )}

      {isCompleted && (
        <div className="mt-auto flex items-center justify-center gap-1.5 rounded bg-green-100/80 py-1.5 text-[10px] font-bold text-green-700">
          <CheckIcon className="w-3.5 h-3.5" />
          Effect Active
        </div>
      )}

      {isLocked && (
        <div className="mt-auto flex items-center justify-center gap-1.5 rounded bg-amber-200/50 py-1.5 text-[10px] text-amber-500">
          <LockIcon className="w-3 h-3" />
          Locked
        </div>
      )}
    </button>
  )
}

// =========== Detail panel ===========

function ResearchDetail({ project }: { project: ResearchProject }) {
  const {
    name, status, building, progress, cashCost, durationHours,
    resourceCost, unlockPct, startedAt, completesAt,
  } = project

  if (!project.id) {
    return (
      <div className="flex items-center justify-center h-40 text-xs text-amber-400 italic">
        Select a project to view details.
      </div>
    )
  }

  const isAvailable = status === 'available'
  const isInProgress = status === 'in_progress'
  const isCompleted = status === 'completed'

  return (
    <div className="space-y-4">
      {/* Header */}
      <div>
        <p className="text-[10px] font-bold uppercase tracking-wider text-amber-500/70">
          {inferCategory(name)} Research
        </p>
        <h3 className="text-lg font-black text-amber-900">{name}</h3>
        {building && (
          <p className="text-[11px] text-amber-600">{building}</p>
        )}
      </div>

      {/* Status badge */}
      <div className="flex items-center gap-2">
        {isAvailable && (
          <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-3 py-1 text-xs font-bold text-green-700">
            Available
          </span>
        )}
        {isInProgress && (
          <span className="inline-flex items-center gap-1 rounded-full bg-blue-100 px-3 py-1 text-xs font-bold text-blue-700">
            Researching...
          </span>
        )}
        {isCompleted && (
          <span className="inline-flex items-center gap-1 rounded-full bg-green-100 px-3 py-1 text-xs font-bold text-green-700">
            <CheckIcon className="w-3.5 h-3.5" /> Completed
          </span>
        )}
      </div>

      {/* Cost breakdown */}
      <div className="rounded-lg bg-amber-100/40 p-3 space-y-1.5">
        <h4 className="text-[10px] font-bold uppercase tracking-wider text-amber-600">Costs</h4>
        {cashCost !== undefined && cashCost > 0 && (
          <div className="flex items-center justify-between text-xs">
            <span className="text-amber-700">Cash</span>
            <span className="font-bold text-green-700">{fmtCash(cashCost)}</span>
          </div>
        )}
        {resourceCost && Object.keys(resourceCost).length > 0 &&
          Object.entries(resourceCost).map(([rid, qty]) => (
            <div key={rid} className="flex items-center justify-between text-xs">
              <span className="text-amber-700">Resource {rid}</span>
              <span className="font-bold text-amber-800">{qty}</span>
            </div>
          ))}
        {durationHours !== undefined && durationHours > 0 && (
          <div className="flex items-center justify-between text-xs">
            <span className="text-amber-700">Duration</span>
            <span className="font-bold text-amber-800">{fmtDuration(durationHours)}</span>
          </div>
        )}
      </div>

      {/* Progress (in-progress only) */}
      {isInProgress && (
        <div className="rounded-lg bg-blue-50/60 p-3 space-y-2">
          <h4 className="text-[10px] font-bold uppercase tracking-wider text-blue-600">Current Progress</h4>
          <div className="h-2 w-full rounded-full bg-blue-200/60 overflow-hidden">
            <div
              className="h-full rounded-full bg-blue-500 transition-all duration-500"
              style={{ width: `${Math.min(progress, 100)}%` }}
            />
          </div>
          <div className="flex items-center justify-between text-xs">
            <span className="text-blue-600 font-semibold">{progress}%</span>
            <span className="text-blue-500">
              <ClockIcon className="inline w-3 h-3 mr-0.5" />
              {timeRemaining(completesAt)}
            </span>
          </div>
          {startedAt && (
            <p className="text-[9px] text-blue-400">
              Started: {new Date(startedAt).toLocaleString()}
            </p>
          )}
        </div>
      )}

      {/* Effect preview */}
      {unlockPct !== undefined && unlockPct > 0 && (
        <div className="rounded-lg bg-amber-50/60 p-3">
          <h4 className="text-[10px] font-bold uppercase tracking-wider text-amber-600 mb-1">Effect</h4>
          <p className="text-xs text-amber-800">
            +{(unlockPct * 100).toFixed(0)}% efficiency
          </p>
        </div>
      )}

      {/* Completed effect */}
      {isCompleted && (
        <div className="rounded-lg bg-green-50/60 p-3 border border-green-200/50">
          <div className="flex items-center gap-2 text-xs text-green-700">
            <CheckIcon className="w-4 h-4" />
            <span className="font-bold">Research completed — effect is active.</span>
          </div>
        </div>
      )}

      {/* Footer hint */}
      <p className="text-[9px] text-amber-400 italic">
        Research unlocks permanent benefits for your entire company.
      </p>
    </div>
  )
}

// =========== Main Page ===========

export function ResearchPage() {
  const [activeCat, setActiveCat] = useState<CategoryId>('production')
  const [selectedId, setSelectedId] = useState<string | null>(null)

  // Primary data: live progress with polling
  const {
    data: progressData,
    isLoading: progressLoading,
    error: progressError,
    refetch: refetchProgress,
  } = useResearchProgress()

  // Fallback: static project list (used when progress endpoint returns empty)
  const {
    data: staticProjects,
    isLoading: staticLoading,
  } = useResearch()

  // Mutations
  const startResearch = useStartResearch()
  const completeResearch = useCompleteResearch()

  // Merge data sources: progress endpoint has full model fields
  const projects: ResearchProject[] = useMemo(() => {
    if (progressData?.projects && progressData.projects.length > 0) {
      return progressData.projects
    }
    if (staticProjects && staticProjects.length > 0) {
      return staticProjects
    }
    return []
  }, [progressData, staticProjects])

  // Filter by category
  const filteredProjects = useMemo(() => {
    return projects.filter((p) => inferCategory(p.name) === activeCat)
  }, [projects, activeCat])

  // Derived: selected project
  const selectedProject = useMemo(() => {
    if (!selectedId) return null
    return projects.find((p) => p.id === selectedId) ?? null
  }, [projects, selectedId])

  // Category counts
  const catCounts = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const p of projects) {
      const cat = inferCategory(p.name)
      counts[cat] = (counts[cat] || 0) + 1
    }
    return counts
  }, [projects])

  // Handlers
  const handleStart = async (projectId: string) => {
    try {
      await startResearch.mutateAsync({ projectId })
      refetchProgress()
    } catch {
      // error handled by mutation state
    }
  }

  const handleComplete = async (projectId: string) => {
    try {
      await completeResearch.mutateAsync(projectId)
      refetchProgress()
    } catch {
      // error handled by mutation state
    }
  }

  const pendingAction = startResearch.isPending || completeResearch.isPending

  // ------ Render ------

  const isLoading = progressLoading && staticLoading
  const hasError = !!progressError && projects.length === 0
  const isEmpty = !isLoading && !hasError && projects.length === 0

  // Full-page loading
  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center">
        <p className="text-amber-600 text-sm italic">Loading research projects...</p>
      </div>
    )
  }

  // Full-page error
  if (hasError) {
    return (
      <div className="h-full flex items-center justify-center">
        <p className="text-red-500 text-sm">Failed to load research data.</p>
      </div>
    )
  }

  // Full-page empty
  if (isEmpty) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="text-center">
          <FlaskIcon className="w-12 h-12 mx-auto text-amber-300 mb-3" />
          <p className="text-amber-600 text-sm">No research projects available.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-7xl p-4 md:p-6">
        {/* ===== Page header ===== */}
        <div className="mb-4">
          <div className="flex items-center gap-2">
            <FlaskIcon className="w-6 h-6 text-blue-600" />
            <h2 className="text-2xl font-black text-amber-950">Research Lab</h2>
          </div>
          <p className="text-xs text-amber-600 mt-0.5">
            Invest in research to unlock powerful upgrades and grow your business.
          </p>
        </div>

        {/* ===== Category tabs ===== */}
        <div className="flex gap-1 mb-5 overflow-x-auto pb-1">
          {CATEGORIES.map((cat) => {
            const isActive = activeCat === cat.id
            const count = catCounts[cat.id] ?? 0
            return (
              <button
                key={cat.id}
                onClick={() => { setActiveCat(cat.id); setSelectedId(null) }}
                className={`
                  flex items-center gap-1.5 rounded-t-lg px-3.5 py-2 text-xs font-bold
                  transition-all duration-150 whitespace-nowrap
                  ${isActive
                    ? 'bg-green-600 text-white shadow-sm'
                    : 'bg-amber-100/70 text-amber-700 hover:bg-amber-200/50'
                  }
                `}
              >
                <span className="w-4 h-4">{cat.icon}</span>
                {cat.label}
                {count > 0 && (
                  <span className={`
                    text-[9px] px-1 rounded
                    ${isActive ? 'bg-green-500 text-white' : 'bg-amber-200/60 text-amber-600'}
                  `}>
                    {count}
                  </span>
                )}
              </button>
            )
          })}
        </div>

        {/* ===== Main content grid ===== */}
        <div className="grid gap-5 lg:grid-cols-[1fr_320px]">
          {/* Left: project card grid */}
          <div>
            {filteredProjects.length === 0 && (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <FlaskIcon className="w-10 h-10 text-amber-300 mb-2" />
                <p className="text-xs text-amber-500">No projects in this category.</p>
              </div>
            )}

            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
              {filteredProjects.map((p) => (
                <ResearchCard
                  key={p.id}
                  project={p}
                  isSelected={selectedId === p.id}
                  onSelect={() => setSelectedId(p.id)}
                  onStart={() => handleStart(p.id)}
                  onComplete={() => handleComplete(p.id)}
                  pendingAction={pendingAction}
                />
              ))}
            </div>
          </div>

          {/* Right: detail panel */}
          <div className="lg:sticky lg:top-4 lg:self-start">
            <div className="rounded-lg border border-amber-200/60 bg-amber-50/60 p-4">
              {selectedProject ? (
                <ResearchDetail project={selectedProject} />
              ) : (
                <div className="flex flex-col items-center justify-center h-40 text-center">
                  <FlaskIcon className="w-8 h-8 text-amber-300 mb-2" />
                  <p className="text-xs text-amber-400 italic">
                    Select a project to view details.
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
