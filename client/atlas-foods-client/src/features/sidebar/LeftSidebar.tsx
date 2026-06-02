import { useUIStore, type ActiveView } from '@/store/ui.store'

const navItems: Array<{ id: ActiveView; label: string; icon: string; iconClass?: string }> = [
  { id: 'map', label: 'Map', icon: '/assets/icons/icon_level_badge_v1.png' },
  { id: 'build', label: 'Build', icon: '/assets/icons/icon_builder_v1.png' },
  { id: 'warehouse', label: 'Warehouse', icon: '/assets/icons/icon_warehouse_v1.png' },
  { id: 'market', label: 'Market', icon: '/assets/icons/icon_market_v1.png' },
  { id: 'contracts', label: 'Contracts', icon: '/assets/icons/icon_contract_v1.png' },
  { id: 'research', label: 'Research', icon: '/assets/icons/icon_timer_v1.png' },
  { id: 'executives', label: 'Executives', icon: '/assets/icons/icon_factory_v1.png' },
  { id: 'finance', label: 'Finance', icon: '/assets/icons/icon_cash_v1.png' },
  { id: 'leaderboard', label: 'Leaderboard', icon: '/assets/icons/icon_level_badge_v1.png' },
]

export function LeftSidebar() {
  const activeView = useUIStore((s) => s.activeView)
  const setActiveView = useUIStore((s) => s.setActiveView)

  return (
    <nav className="sidebar flex flex-col bg-[#f5e6c8] border-r-2 border-amber-700/30 overflow-y-auto">
      {navItems.map((item) => (
        <button
          key={item.id}
          onClick={() => setActiveView(item.id)}
          className={`
            flex flex-col items-center justify-center gap-1 py-4 px-2
            border-b border-amber-700/10 transition-all duration-150
            ${activeView === item.id
              ? 'bg-amber-200/70 text-amber-900 font-semibold shadow-inner'
              : 'text-amber-800/70 hover:bg-amber-100/50 hover:text-amber-900'
            }
          `}
          style={{ height: activeView === item.id ? '68px' : '80px' }}
        >
          <img
            src={item.icon}
            alt={item.label}
            className="w-9 h-9 object-contain"
            loading="lazy"
          />
          <span className="text-[11px] uppercase tracking-wider">{item.label}</span>
        </button>
      ))}

      {/* Company Reputation Card */}
      <div className="mt-auto p-3 mx-2 mb-3 bg-amber-900/10 rounded-lg border border-amber-700/20">
        <div className="text-[9px] uppercase tracking-widest text-amber-700/60 text-center mb-1">
          Company Reputation
        </div>
        <div className="flex items-center justify-center gap-1">
          <svg className="w-3 h-3 text-yellow-600" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 1.944A11.954 11.954 0 012.166 5C2.056 5.647 2 6.319 2 7c0 5.225 3.34 9.67 8 11.317C14.66 16.67 18 12.225 18 7c0-.682-.057-1.353-.166-2A11.954 11.954 0 0110 1.944zM11 14a1 1 0 11-2 0 1 1 0 012 0zm0-7a1 1 0 10-2 0v3a1 1 0 102 0V7z" clipRule="evenodd" />
          </svg>
          <span className="text-xs font-bold text-amber-800">5</span>
        </div>
        <div className="text-[9px] text-amber-700/50 text-center mt-0.5">Trusted Partner</div>
      </div>
    </nav>
  )
}
