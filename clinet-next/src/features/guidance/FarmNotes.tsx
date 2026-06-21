import { useMemo, useState, useEffect } from 'react'
import { useBuildings } from '@/api/hooks/buildings.hooks'
import { useWarehouse } from '@/api/hooks/warehouse.hooks'
import { useProductionJobs } from '@/api/hooks/production.hooks'
import { useCompany, useCompleteTutorial } from '@/api/hooks/company.hooks'
import { resourceIcon } from '@/game/resources'
import { useUIStore } from '@/store/ui.store'
import { useTranslation } from 'react-i18next'

type GuideArrow = 'left' | 'right' | 'down'
type GuideCardPosition = 'leftNav' | 'centerTop' | 'rightPanel' | 'marketLeft'

interface GuideNote {
  title: string
  body: string
  resourceId: number
  arrow: GuideArrow
  card?: GuideCardPosition
}

// Post-tutorial notes (Market Unlocked, Try The Chain, Keep Expanding, etc.)
// are kept in git history but currently not rendered after tutorial completion.

export function FarmNotes() {
  const { t } = useTranslation()
  const { data: buildingsData } = useBuildings()
  const { data: warehouse } = useWarehouse()
  const { data: jobsData } = useProductionJobs()
  const { data: companyData } = useCompany()
  const completeTutorial = useCompleteTutorial()
  const [showTips, setShowTips] = useState(true)
  const [showCelebration, setShowCelebration] = useState(false)
  const [celebrationDone, setCelebrationDone] = useState(
    companyData?.levelInfo?.tutorialCompleted ?? false,
  )
  const activeView = useUIStore((s) => s.activeView)
  const selectedBuildingId = useUIStore((s) => s.selectedBuildingId)
  const placementBuildingId = useUIStore((s) => s.placementBuildingId)

  // Read tutorial setting from localStorage (same key as SettingsPage)
  useEffect(() => {
    try {
      const raw = localStorage.getItem('atlas_foods_settings')
      if (raw) {
        const parsed = JSON.parse(raw)
        setShowTips(parsed.showTutorialTips !== false)
      }
    } catch { /* ignore */ }
  }, [])

  const buildings = Array.isArray(buildingsData) ? buildingsData : []
  const placed = buildings.filter((b) => b.placed !== false)
  const unplaced = buildings.filter((b) => b.placed === false)
  const jobs = jobsData ?? []
  const activeJobs = jobs.filter((job) => job.status !== 'claimed')
  const collectedOnce = jobs.some((job) => job.status === 'claimed' || (job.claimedAmount ?? 0) > 0)
  const collectable = jobs.reduce((sum, job) => sum + (job.claimableAmount ?? 0), 0)
  const selectedJobs = selectedBuildingId
    ? jobs.filter((job) => job.buildingId === selectedBuildingId && job.status !== 'claimed')
    : []
  const selectedActiveJobs = selectedJobs.filter((job) => job.status !== 'claimed')
  const selectedCollectable = selectedJobs.reduce((sum, job) => sum + (job.claimableAmount ?? 0), 0)
  const rawGoods = (warehouse?.inventory ?? [])
    .filter((item) => [1, 2, 6, 12].includes(item.resourceId))
    .reduce((sum, item) => sum + item.quantity, 0)
  const level = companyData?.levelInfo?.level ?? 1

  // Auto-dismiss celebration after 6 seconds
  useEffect(() => {
    if (!showCelebration) return
    const timer = setTimeout(() => setShowCelebration(false), 6000)
    return () => clearTimeout(timer)
  }, [showCelebration])

  // Detect tutorial completion: first harvest done and no active jobs left
  useEffect(() => {
    if (collectedOnce && activeJobs.length === 0 && !celebrationDone) {
      setShowCelebration(true)
      setCelebrationDone(true)
      // Persist to backend so it survives browser clears
      completeTutorial.mutate()
    }
  }, [collectedOnce, activeJobs.length, celebrationDone])

  // Pre-tutorial: early guidance before first complete harvest cycle
  const note = useMemo<GuideNote | null>(() => {
    if (collectedOnce && activeJobs.length === 0) {
      // Tutorial just hit the milestone — celebration shown above, return no card here
      return null
    }

    if (activeView === 'build') {
      if (unplaced.length > 0) {
        return {
          title: t('tutorial.scrollDown'),
          body: t('tutorial.scrollDownBody'),
          resourceId: 12,
          arrow: 'down',
          card: 'centerTop',
        }
      }
      if (placed.length > 0) {
        return {
          title: t('tutorial.returnToMap'),
          body: t('tutorial.returnToMapBody'),
          resourceId: 1,
          arrow: 'left',
          card: 'leftNav',
        }
      }
      return {
          title: t('tutorial.buyBuilding'),
          body: t('tutorial.buyBuildingBody'),
        resourceId: 1,
        arrow: 'right',
        card: 'marketLeft',
      }
    }

    if (activeView === 'market') {
      return {
          title: t('tutorial.marketTutorial'),
          body: t('tutorial.marketTutorialBody'),
        resourceId: 1,
        arrow: 'right',
        card: 'marketLeft',
      }
    }

    if (activeView !== 'map') {
      return null
    }

    if (placed.length === 0 && unplaced.length === 0) {
      return {
          title: t('tutorial.startHere'),
          body: t('tutorial.startHereBody'),
        resourceId: 1,
        arrow: 'left',
        card: 'leftNav',
      }
    }
    if (unplaced.length > 0) {
      if (placementBuildingId) {
        return {
          title: t('tutorial.placeItHere'),
          body: t('tutorial.placeItHereBody'),
          resourceId: 12,
          arrow: 'down',
          card: 'centerTop',
        }
      }
      return {
          title: t('tutorial.backToBuild'),
          body: t('tutorial.backToBuildBody'),
        resourceId: 12,
        arrow: 'left',
        card: 'leftNav',
      }
    }

    if (selectedBuildingId) {
      if (selectedCollectable > 0) {
        return {
          title: t('tutorial.collectIt'),
          body: t('tutorial.collectItBody'),
          resourceId: 1,
          arrow: 'right',
          card: 'rightPanel',
        }
      }
      if (selectedActiveJobs.length > 0) {
        return {
          title: t('tutorial.waitForYield'),
          body: t('tutorial.waitForYieldBody'),
          resourceId: selectedActiveJobs[0]?.resourceId ?? 1,
          arrow: 'right',
          card: 'rightPanel',
        }
      }
      if (collectedOnce) {
        return null
      }
      return {
          title: t('tutorial.setProduction'),
          body: t('tutorial.setProductionBody'),
        resourceId: 1,
        arrow: 'right',
        card: 'rightPanel',
      }
    }

    if (collectedOnce && activeJobs.length === 0) {
      return null
    }

    if (placed.length > 0) {
      return {
          title: t('tutorial.openBuilding'),
          body: t('tutorial.openBuildingBody'),
        resourceId: 1,
        arrow: 'down',
        card: 'centerTop',
      }
    }

    if (collectable > 0) {
      return {
          title: t('tutorial.collectNow'),
          body: t('tutorial.collectNowBody', { count: collectable.toLocaleString() }),
        resourceId: 1,
        arrow: 'down',
      }
    }
    if (jobs.length > 0) {
      return {
          title: t('tutorial.watchProduction'),
          body: t('tutorial.watchProductionBody'),
        resourceId: jobs[0]?.resourceId ?? 1,
        arrow: 'down',
      }
    }

    return null
  }, [activeJobs.length, activeView, collectable, collectedOnce, jobs, level, placed.length, placementBuildingId, rawGoods, selectedActiveJobs, selectedBuildingId, selectedCollectable, t, unplaced.length])

  if (!showTips) return null

  // Celebration card after first complete harvest cycle
  if (showCelebration) {
    return (
      <aside className="pointer-events-none fixed z-40 max-w-[340px] left-1/2 top-[106px] -translate-x-1/2">
        <div className="relative rounded-xl border-2 border-amber-500/70 bg-gradient-to-br from-amber-50 to-amber-100/95 px-6 py-5 shadow-2xl shadow-amber-950/20 backdrop-blur">
          <div className="absolute -inset-1 -z-10 rounded-[14px] border-2 border-amber-300/70 farm-guide-pulse" />
          <div className="flex flex-col items-center text-center gap-2">
            <svg className="w-10 h-10 text-amber-500 farm-celebration-sparkle" viewBox="0 0 24 24" fill="none">
              <path d="M12 2l1.5 6.5L20 9l-5 4.5 1.5 7L12 16l-5.5 4.5L8 14.5 3 10l6.5-.5L12 2z" fill="currentColor" />
            </svg>
            <div className="text-[11px] font-black uppercase tracking-[0.18em] text-amber-700">{t('tutorial.congratulations')}</div>
            <div className="text-lg font-black leading-tight text-amber-950">{t('tutorial.congratulationsBody')}</div>
            <p className="text-xs font-semibold leading-snug text-amber-800">
              You know the basics. Explore the market, expand your chains, and discover what works.
            </p>
          </div>
        </div>
        <svg className="farm-guide-arrow absolute left-1/2 top-[calc(100%+8px)] h-20 w-16 -translate-x-1/2 text-amber-500/60 drop-shadow-lg" viewBox="0 0 48 64" fill="none">
          <path d="M24 6v40" stroke="currentColor" strokeWidth="6" strokeLinecap="round" />
          <path d="M10 38l14 20 14-20" fill="currentColor" />
        </svg>
      </aside>
    )
  }

  // Tutorial permanently done after celebration — never show cards again
  if (celebrationDone) return null

  if (!note) return null

  const cardClass = note.card === 'rightPanel'
    ? 'right-[320px] top-[132px]'
    : note.card === 'marketLeft'
    ? 'left-[320px] top-[132px]'
    : note.card === 'leftNav'
    ? 'left-[128px] top-[108px]'
    : 'left-1/2 top-[106px] -translate-x-1/2'

  return (
    <aside className={`pointer-events-none fixed z-40 max-w-[340px] ${cardClass}`}>
      <div className="relative rounded-xl border-2 border-amber-500/70 bg-amber-50/95 px-4 py-3 shadow-2xl shadow-amber-950/20 backdrop-blur">
        <div className="absolute -inset-1 -z-10 rounded-[14px] border-2 border-amber-300/70 farm-guide-pulse" />
        <div className="flex items-start gap-3">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg border border-amber-300/70 bg-white/80">
            <img src={resourceIcon(note.resourceId)} alt="" className="h-9 w-9 object-contain" />
          </div>
          <div className="min-w-0">
            <div className="text-[11px] font-black uppercase tracking-[0.18em] text-amber-700">New Farmer</div>
            <div className="mt-0.5 text-lg font-black leading-tight text-amber-950">{note.title}</div>
            <p className="mt-1 text-xs font-semibold leading-snug text-amber-800">{note.body}</p>
          </div>
        </div>
      </div>
      <FarmGuideArrow direction={note.arrow} />
    </aside>
  )
}

function FarmGuideArrow({ direction }: { direction: GuideArrow }) {
  if (direction === 'left') {
    return (
      <svg className="farm-guide-arrow-left absolute -left-20 top-3 h-16 w-20 text-amber-600 drop-shadow-lg" viewBox="0 0 96 72" fill="none">
        <path d="M88 36H18" stroke="currentColor" strokeWidth="12" strokeLinecap="round" />
        <path d="M30 12L8 36l22 24" fill="currentColor" />
      </svg>
    )
  }

  if (direction === 'right') {
    return (
      <svg className="farm-guide-arrow-right absolute -right-20 top-3 h-16 w-20 text-amber-600 drop-shadow-lg" viewBox="0 0 96 72" fill="none">
        <path d="M8 36h70" stroke="currentColor" strokeWidth="12" strokeLinecap="round" />
        <path d="M66 12l22 24-22 24" fill="currentColor" />
      </svg>
    )
  }

  return (
    <svg className="farm-guide-arrow absolute left-1/2 top-[calc(100%+8px)] h-28 w-24 -translate-x-1/2 text-amber-600 drop-shadow-lg" viewBox="0 0 96 128" fill="none">
      <path d="M48 8v82" stroke="currentColor" strokeWidth="14" strokeLinecap="round" />
      <path d="M16 76l32 44 32-44" fill="currentColor" />
    </svg>
  )
}
