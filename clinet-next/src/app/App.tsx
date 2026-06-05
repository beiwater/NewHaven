import { useUIStore, type ActiveView } from '@/store/ui.store'
import { MarketPage } from '@/features/market/MarketPage'
import { WarehousePage } from '@/features/warehouse/WarehousePage'
import { ProductionQueue } from '@/features/production/ProductionQueue'
import { BuildView } from '@/features/buildings/BuildView'

const NAV_ITEMS: { id: ActiveView; label: string }[] = [
  { id: 'market', label: 'Market' },
  { id: 'warehouse', label: 'Warehouse' },
  { id: 'production', label: 'Production' },
  { id: 'build', label: 'Build' },
]

function PageContent() {
  const activeView = useUIStore((s) => s.activeView)

  switch (activeView) {
    case 'market':
      return <MarketPage />
    case 'warehouse':
      return <WarehousePage />
    case 'production':
      return <ProductionQueue />
    case 'build':
      return <BuildView />
    default:
      return <MarketPage />
  }
}

export function App() {
  const setActiveView = useUIStore((s) => s.setActiveView)
  const activeView = useUIStore((s) => s.activeView)

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
              className={`px-3 py-1.5 text-xs font-semibold rounded-md transition-colors ${
                activeView === item.id
                  ? 'bg-amber-600 text-white'
                  : 'text-amber-200 hover:bg-amber-700 hover:text-white'
              }`}
            >
              {item.label}
            </button>
          ))}
        </nav>
      </header>

      {/* Page content */}
      <main className="flex-1 overflow-hidden">
        <PageContent />
      </main>
    </div>
  )
}
