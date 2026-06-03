import { Suspense, lazy, useState } from 'react'
import { ErrorBoundary } from '@/app/ErrorBoundary'
import { useUIStore } from '@/store/ui.store'
import { useMarketWebSocket, useProductionWebSocket } from '@/api/websocket'
import { MobileTopBar } from './MobileTopBar'
import { MobileResourceBar } from './MobileResourceBar'
import { MobileBottomNav } from './MobileBottomNav'
import { MobileDrawer } from './MobileDrawer'
import { MobileOrderBoard } from './MobileOrderBoard'
import { MobileBuildingSummary } from './MobileBuildingSummary'
import { MobileFactoryQueue } from './MobileFactoryQueue'
import { ChatPanel } from '@/features/chat/ChatPanel'
import { PowerPanel } from '@/features/powerups/PowerPanel'
import { MarketPage } from '@/features/market/MarketPage'
import { InventoryBar } from '@/features/inventory/InventoryBar'
import { BuildView } from '@/features/buildings/BuildView'
import { ContractList } from '@/features/contracts/ContractList'
import { ExecutivePage } from '@/features/executives/ExecutivePage'
import { FinancialPage } from '@/features/financial/FinancialPage'
import { ResearchPage } from '@/features/research/ResearchPage'
import { LeaderboardPage } from '@/features/leaderboard/LeaderboardPage'
import { SettingsPage } from '@/features/settings/SettingsPage'
import { InspectPage } from '@/features/inspect/InspectPage'

const GameCanvas = lazy(() => import('@/game/GameCanvas'))

function MobilePageContent() {
  const activeView = useUIStore((s) => s.activeView)

  switch (activeView) {
    case 'market':
      return <MarketPage />
    case 'warehouse':
      return <InventoryBar />
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

export function MobileLayout() {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const activeView = useUIStore((s) => s.activeView)
  const chatOpen = useUIStore((s) => s.chatOpen)
  const setChatOpen = useUIStore((s) => s.setChatOpen)
  const isMapView = activeView === 'map'

  useMarketWebSocket()
  useProductionWebSocket()

  return (
    <div className="mobile-layout">
      {/* Top bar */}
      <MobileTopBar onMenuClick={() => setDrawerOpen(true)} />

      {/* Resource bar */}
      <MobileResourceBar />

      {/* Map area / Page content */}
      <div className="mobile-map">
        {isMapView ? (
          <ErrorBoundary fallback={
            <div className="flex items-center justify-center h-full bg-amber-100">
              <div className="text-center p-4">
                <p className="text-amber-700 font-semibold text-sm">Map loading failed</p>
                <p className="text-amber-500 text-xs mt-1">WebGL may not be supported</p>
              </div>
            </div>
          }>
            <Suspense fallback={
              <div className="flex items-center justify-center h-full bg-amber-100">
                <div className="text-amber-600 text-sm animate-pulse">Loading map...</div>
              </div>
            }>
              <GameCanvas />
            </Suspense>
          </ErrorBoundary>
        ) : (
          <div className="h-full overflow-y-auto">
            <MobilePageContent />
          </div>
        )}
      </div>

      {/* Bottom panels (map view only) */}
      {isMapView && (
        <div className="mobile-panels">
          <MobileBuildingSummary />
          <MobileOrderBoard />
          <MobileFactoryQueue />
        </div>
      )}

      {/* Chat fab */}
      {!chatOpen && (
        <button
          onClick={() => setChatOpen(true)}
          className="fixed bottom-[124px] right-3 z-40 w-9 h-9 bg-amber-800 hover:bg-amber-900 text-white rounded-full shadow-lg flex items-center justify-center transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        </button>
      )}
      <ChatPanel />
      <PowerPanel />

      {/* Bottom nav */}
      <MobileBottomNav />

      {/* Drawer */}
      <MobileDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} />
    </div>
  )
}
