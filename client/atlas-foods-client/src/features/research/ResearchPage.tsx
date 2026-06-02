import { useState, useMemo, useEffect } from 'react'
import {
  useResearch,
  useResearchProgress,
  useStartResearch,
  useCompleteResearch,
} from '@/api/research.api'
import type { ResearchProject } from '@/api/research.api'

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

// =========== Icons (inline SVG, no asset imports) ===========

function FlaskIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
    </svg>
  )
}

function GearIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  )
}

function PriceTagIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 7h.01M7 3h5a2 2 0 012 2v5l9 9-3 3-9-9H7a2 2 0 01-2-2V5a2 2 0 012-2z" />
    </svg>
  )
}

function TruckIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 7h12a1 1 0 011 1v8a1 1 0 01-1 1h-1M8 7H4a1 1 0 00-1 1v8a1 1 0 001 1h1m10-9v4h3l-3-4zm-6 9a2 2 0 100-4 2 2 0 000 4zm8 0a2 2 0 100-4 2 2 0 000 4z" />
    </svg>
  )
}

function CutleryIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
    </svg>
  )
}

function BankIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 21h18M3 10h18M5 6l7-3 7 3M4 10v11m16-11v11M8 14v4m4-4v4m4-4v4" />
    </svg>
  )
}

function StarIcon({ className = 'w-3 h-3' }: { className?: string }) {
  return (
    <svg className={className} fill="currentColor" viewBox="0 0 20 20">
      <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
    </svg>
  )
}

function CheckIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
    </svg>
  )
}

function LockIcon({ className = 'w-4 h-4' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
    </svg>
  )
}

function ClockIcon({ className = 'w-3 h-3' }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  )
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

  // Selected project
  const selectedProject = useMemo(() => {
    if (!selectedId) return null
    return projects.find((p) => p.id === selectedId) ?? null
  }, [projects, selectedId])

  // When projects change, ensure selected ID is still valid
  useEffect(() => {
    if (selectedId && !projects.some((p) => p.id === selectedId)) {
      setSelectedId(null)
    }
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
