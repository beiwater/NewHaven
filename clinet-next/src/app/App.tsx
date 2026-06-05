import { useUIStore, type ActiveView } from '@/store/ui.store'
import { MarketPage } from '@/features/market/MarketPage'
import { WarehousePage } from '@/features/warehouse/WarehousePage'
import { ProductionQueue } from '@/features/production/ProductionQueue'
import { BuildView } from '@/features/buildings/BuildView'
import { ResearchPage } from '@/features/research/ResearchPage'
import { FinancialPage } from '@/features/financial/FinancialPage'
import { ExecutivePage } from '@/features/executives/ExecutivePage'
import { ChatPanel } from '@/features/chat/ChatPanel'
import { LeaderboardPage } from '@/features/leaderboard/LeaderboardPage'
import { PowerPanel } from '@/features/powerups/PowerPanel'
import { ContractList } from '@/features/contracts/ContractList'

const NAV_ITEMS: { id: ActiveView; label: string }[] = [
  { id: 'market', label: 'Market' },
  { id: 'warehouse', label: 'Warehouse' },
  { id: 'production', label: 'Production' },
  { id: 'build', label: 'Build' },
  { id: 'research', label: 'Research' },
  { id: 'finance', label: 'Finance' },
  { id: 'executives', label: 'Executives' },
  { id: 'leaderboard', label: 'Leaderboard' },
  { id: 'contracts', label: 'Contracts' },
]

function PageContent() {
  const activeView = useUIStore((s) => s.activeView)
  const chatOpen = useUIStore((s) => s.chatOpen)
  const powerupOpen = useUIStore((s) => s.powerupOpen)

  const main = (() => {
    switch (activeView) {
      case 'market': return <MarketPage />
      case 'warehouse': return <WarehousePage />
      case 'production': return <ProductionQueue />
      case 'build': return <BuildView />
      case 'research': return <ResearchPage />
      case 'finance': return <FinancialPage />
      case 'executives': return <ExecutivePage />
      case 'leaderboard': return <LeaderboardPage />
      case 'contracts': return <ContractList />
      default: return <MarketPage />
    }
  })()

  return (
    <div className="flex-1 flex relative overflow-hidden">
      {main}
      {chatOpen && (
        <div className="absolute right-0 top-0 bottom-0 w-80 border-l-2 border-amber-700/30 bg-amber-50 shadow-xl z-40 overflow-y-auto">
          <ChatPanel />
        </div>
      )}
      {powerupOpen && (
        <div className="absolute right-0 top-0 bottom-0 w-80 border-l-2 border-amber-700/30 bg-amber-50 shadow-xl z-40 overflow-y-auto">
          <PowerPanel />
        </div>
      )}
    </div>
  )
}

export function App() {
  const setActiveView = useUIStore((s) => s.setActiveView)
  const activeView = useUIStore((s) => s.activeView)
  const setChatOpen = useUIStore((s) => s.setChatOpen)
  const chatOpen = useUIStore((s) => s.chatOpen)
  const setPowerupOpen = useUIStore((s) => s.setPowerupOpen)
  const powerupOpen = useUIStore((s) => s.powerupOpen)

  return (
    <div className="min-h-screen bg-amber-50 flex flex-col">
      {/* Top nav bar */}
      <header className="flex items-center justify-between px-4 py-2 bg-amber-800 text-white shadow-md">
        <h1 className="text-sm font-bold uppercase tracking-wider">New Haven</h1>
        <nav className="flex gap-1">
          {NAV_ITEMS.map((item) => (
            <button
              key={item.id}
              onClick={() => setActiveView(item.id)}
              className={`px-2.5 py-1.5 text-[11px] font-semibold rounded-md transition-colors ${
                activeView === item.id
                  ? 'bg-amber-600 text-white'
                  : 'text-amber-200 hover:bg-amber-700 hover:text-white'
              }`}
            >
              {item.label}
            </button>
          ))}
          {/* Chat toggle */}
          <button
            onClick={() => setChatOpen(!chatOpen)}
            className={`px-2.5 py-1.5 text-[11px] font-semibold rounded-md transition-colors ${
              chatOpen ? 'bg-blue-600 text-white' : 'text-amber-200 hover:bg-amber-700 hover:text-white'
            }`}
          >
            Chat
          </button>
          {/* Power-up toggle */}
          <button
            onClick={() => setPowerupOpen(!powerupOpen)}
            className={`px-2.5 py-1.5 text-[11px] font-semibold rounded-md transition-colors ${
              powerupOpen ? 'bg-purple-600 text-white' : 'text-amber-200 hover:bg-amber-700 hover:text-white'
            }`}
          >
            Power-ups
          </button>
        </nav>
      </header>

      {/* Page content */}
      <main className="flex-1 overflow-hidden">
        <PageContent />
      </main>
    </div>
  )
}
