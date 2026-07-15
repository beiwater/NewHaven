import { lazy, Suspense } from 'react'
import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { AuthGate } from '@/features/auth/AuthGate'
import { TopBar } from '@/features/topbar/TopBar'
import { LeftSidebar } from '@/features/sidebar/LeftSidebar'
import { BuildingPanel } from '@/features/buildings/BuildingPanel'
import { MarketTicker } from '@/features/market/MarketTicker'
import { MarketPage } from '@/features/market/MarketPage'
import { InventoryBar } from '@/features/inventory/InventoryBar'
import { ContractList } from '@/features/contracts/ContractList'
import { ChatPanel } from '@/features/chat/ChatPanel'
import { ReportPanel } from '@/features/report/ReportPanel'
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
import { ChatPage } from '@/features/chat/ChatPage'
import { InspectPage } from '@/features/inspect/InspectPage'
import { FarmNotes } from '@/features/guidance/FarmNotes'
import { GameWikiPage } from '@/features/guidance/GameWikiPage'
import { ProductionQueue } from '@/features/production/ProductionQueue'
import { useIsMobile } from '@/hooks/useIsMobile'
import { MobileLayout } from '@/features/mobile'
import { StoryGate } from '@/features/story/StoryGate'
import { useAudio } from '@/audio/useAudio'

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
    case 'chat':
      return <ChatPage />
    case 'wiki':
      return <GameWikiPage />

    case 'inspect':
      return <InspectPage />
    case 'map':
    default:
      return null
  }
}

function MapSlot() {
  const { t } = useTranslation()
  return (
    <ErrorBoundary fallback={
      <div className="map flex items-center justify-center bg-amber-100">
        <div className="text-center p-8">
          <p className="text-amber-700 font-semibold">{t('common.mapLoadingFailed')}</p>
          <p className="text-amber-500 text-xs mt-1">{t('common.webglNotSupported')}</p>
        </div>
      </div>
    }>
      <Suspense fallback={
        <div className="map flex items-center justify-center bg-amber-100">
          <div className="text-amber-600 text-sm animate-pulse">{t('common.loadingMap')}</div>
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
  const { playMusic, unlockAudio } = useAudio()
  const unlocked = useRef(false)

  useMarketWebSocket()
  useProductionWebSocket()

  // Unlock audio context on first user interaction
  const handleFirstInteraction = () => {
    if (!unlocked.current) {
      unlocked.current = true
      unlockAudio()
    }
  }

  // BGM scene switching
  useEffect(() => {
    switch (activeView) {
      case 'market':
        playMusic('bgm_market')
        break
      case 'build':
      case 'map':
        playMusic('bgm_harbor_town')
        break
      case 'warehouse':
      case 'chain':
        playMusic('bgm_harbor_town')
        break
      case 'production':
        playMusic('bgm_harbor_town')
        break
      case 'contracts':
        playMusic('bgm_harbor_town')
        break
      case 'research':
      case 'executives':
        playMusic('bgm_harbor_town')
        break
      case 'finance':
      case 'leaderboard':
      case 'settings':
      case 'inspect':
	  case 'wiki':
        playMusic('bgm_harbor_town')
        break
      default:
        playMusic('bgm_harbor_town')
    }
  }, [activeView, playMusic])

  return (
    <div
      className={`game-layout ${isMapView ? 'map-mode' : 'page-mode'}`}
      onClick={handleFirstInteraction}
      onTouchStart={handleFirstInteraction}
      onKeyDown={handleFirstInteraction}
    >
      <TopBar />
      <LeftSidebar />
      <div className="map">
        {isMapView ? <MapSlot /> : <PageContent />}
      </div>
      <FarmNotes />
      {isMapView ? <BuildingPanel /> : <div className="right-panel page-spacer" />}
      <MarketTicker />
      {!chatOpen && activeView !== 'chat' && (
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
      <ReportPanel />
      <PowerPanel />
    </div>
  )
}


export function App() {
  const isMobile = useIsMobile()
  return (
    <ErrorBoundary>
      <AuthGate>
        <StoryGate>
          {isMobile ? <MobileLayout /> : <GameLayout />}
        </StoryGate>
      </AuthGate>
    </ErrorBoundary>
  )
}
