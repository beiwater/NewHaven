import { lazy, Suspense } from 'react'
import '@/i18n'
import { TopBar } from '@/features/topbar/TopBar'
import { LeftSidebar } from '@/features/sidebar/LeftSidebar'
import { BuildingPanel } from '@/features/buildings/BuildingPanel'
import { MarketTicker } from '@/features/market/MarketTicker'
import { MarketPage } from '@/features/market/MarketPage'
import { InventoryBar } from '@/features/inventory/InventoryBar'
import { ContractList } from '@/features/contracts/ContractList'
import { ChatPanel } from '@/features/chat/ChatPanel'
import { PowerPanel } from '@/features/powerups/PowerPanel'
import { BuildView } from '@/features/buildings/BuildView'
import { useUIStore } from '@/store/ui.store'
import { ExecutivePage } from '@/features/executives/ExecutivePage'
import { FinancialPage } from '@/features/financial/FinancialPage'
import { ResearchPage } from '@/features/research/ResearchPage'
import { ErrorBoundary } from './ErrorBoundary'
import { LeaderboardPage } from '@/features/leaderboard/LeaderboardPage'
import { InspectPage } from '@/features/inspect/InspectPage'
import { FarmNotes } from '@/features/guidance/FarmNotes'
import { ProductionQueue } from '@/features/production/ProductionQueue'

const MapSlot = lazy(() => import('@/features/map/MapSlot'))

function PageContent() {
  const activeView = useUIStore((s) => s.activeView)

  switch (activeView) {
    case 'market': return <MarketPage />
    case 'warehouse':
    case 'chain': return <InventoryBar />
    case 'production': return <ProductionQueue />
    case 'build': return <BuildView />
    case 'contracts': return <ContractList />
    case 'research': return <ResearchPage />
    case 'executives': return <ExecutivePage />
    case 'finance': return <FinancialPage />
    case 'leaderboard': return <LeaderboardPage />
    case 'inspect': return <InspectPage />
    default: return (
      <div className="flex items-center justify-center h-full text-amber-600 text-sm p-8">
        Select a page from the sidebar
      </div>
    )
  }
}

function GameLayout() {
  const activeView = useUIStore((s) => s.activeView)
  const setChatOpen = useUIStore((s) => s.setChatOpen)
  const chatOpen = useUIStore((s) => s.chatOpen)
  const isMapView = activeView === 'map'

  return (
    <div
      className={`game-layout ${isMapView ? 'map-mode' : 'page-mode'}`}
    >
      <TopBar />
      <LeftSidebar />
      <div className="map">
        {isMapView ? (
          <Suspense fallback={<div className="p-8 text-amber-600">Loading map...</div>}>
            <MapSlot />
          </Suspense>
        ) : (
          <PageContent />
        )}
      </div>
      <FarmNotes />
      {isMapView ? <BuildingPanel /> : <div className="right-panel page-spacer" />}
      <MarketTicker />

      {!chatOpen && (
        <button
          onClick={(e) => {
            e.stopPropagation()
            setChatOpen(true)
          }}
          className="fixed bottom-[110px] right-4 z-50 w-10 h-10 bg-amber-800 hover:bg-amber-900 text-white rounded-full shadow-lg flex items-center justify-center transition-colors"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        </button>
      )}
      <ChatPanel />
      <PowerPanel />
    </div>
  )
}

export function App() {
  return (
    <ErrorBoundary>
      <GameLayout />
    </ErrorBoundary>
  )
}
