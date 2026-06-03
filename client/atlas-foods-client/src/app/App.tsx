import { lazy, Suspense } from 'react'
import { AuthGate } from '@/features/auth/AuthGate'
import { TopBar } from '@/features/topbar/TopBar'
import { LeftSidebar } from '@/features/sidebar/LeftSidebar'
import { BuildingPanel } from '@/features/buildings/BuildingPanel'
import { MarketTicker } from '@/features/market/MarketTicker'
import { MarketPage } from '@/features/market/MarketPage'
import { InventoryBar } from '@/features/inventory/InventoryBar'
import { ContractList } from '@/features/contracts/ContractList'
import { ChatPanel } from '@/features/chat/ChatPanel'
import { PowerPanel } from '@/features/powerups/PowerPanel'
import { useMarketWebSocket, useProductionWebSocket } from '@/api/websocket'
import { BuildView } from '@/features/buildings/BuildView'
import { useUIStore } from '@/store/ui.store'
import { ExecutivePage } from '@/features/executives/ExecutivePage'
import { FinancialPage } from '@/features/financial/FinancialPage'
import { ResearchPage } from '@/features/research/ResearchPage'
import { ErrorBoundary } from './ErrorBoundary'
import { LeaderboardPage } from '@/features/leaderboard/LeaderboardPage'
import { SettingsPage } from '@/features/settings/SettingsPage'
import { InspectPage } from '@/features/inspect/InspectPage'
import { FarmNotes } from '@/features/guidance/FarmNotes'
import { ProductionQueue } from '@/features/production/ProductionQueue'
import { useIsMobile } from '@/hooks/useIsMobile'
import { MobileLayout } from '@/features/mobile'

// Lazy-load PixiJS game canvas so it doesn't block React mount
const GameCanvas = lazy(() => import('@/game/GameCanvas'))

function PageContent() {
  const activeView = useUIStore((s) => s.activeView)

  switch (activeView) {
    case 'market':
      return <MarketPage />
    case 'chain':
    case 'warehouse':
      return <InventoryBar />
    case 'production':
      return <ProductionQueue />
    case 'build':
      return <BuildView />
    case 'contracts':
      return <ContractList />
    case 'executives':
      return <ExecutivePage />
    case 'finance':
      return <FinancialPage />
    case 'research':
      return <ResearchPage />
    case 'leaderboard':
      return <LeaderboardPage />
    case 'settings':
      return <SettingsPage />
    case 'inspect':
      return <InspectPage />
    case 'map':
    default:
      return null
  }
}

function MapSlot() {
  return (
    <ErrorBoundary fallback={
      <div className="map flex items-center justify-center bg-amber-100">
        <div className="text-center p-8">
          <p className="text-amber-700 font-semibold">Map loading failed</p>
          <p className="text-amber-500 text-xs mt-1">WebGL may not be supported</p>
        </div>
      </div>
    }>
      <Suspense fallback={
        <div className="map flex items-center justify-center bg-amber-100">
          <div className="text-amber-600 text-sm animate-pulse">Loading map...</div>
        </div>
      }>
        <GameCanvas />
      </Suspense>
    </ErrorBoundary>
  )
}

function GameLayout() {
  const activeView = useUIStore((s) => s.activeView)
  const setChatOpen = useUIStore((s) => s.setChatOpen)
  const chatOpen = useUIStore((s) => s.chatOpen)
  const isMapView = activeView === 'map'

  useMarketWebSocket()
  useProductionWebSocket()

  return (
    <div className={`game-layout ${isMapView ? 'map-mode' : 'page-mode'}`}>
      <TopBar />
      <LeftSidebar />
      <div className="map">
        {isMapView ? <MapSlot /> : <PageContent />}
      </div>
      <FarmNotes />
      {isMapView ? <BuildingPanel /> : <div className="right-panel page-spacer" />}
      <MarketTicker />
      {!chatOpen && (
        <button
          onClick={() => setChatOpen(true)}
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
  const isMobile = useIsMobile()
  return (
    <ErrorBoundary>
      <AuthGate>
        {isMobile ? <MobileLayout /> : <GameLayout />}
      </AuthGate>
    </ErrorBoundary>
  )
}
