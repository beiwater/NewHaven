import { useUIStore } from '@/store/ui.store'
import { useCompany } from '@/api/company.api'
import { Icon } from '@/features/ui/Icon'

export function MobileTopBar({ onMenuClick }: { onMenuClick: () => void }) {
  const { data: companyData } = useCompany()
  const companyName = companyData?.authCompany?.company ?? 'Mellow Acres Co.'
  const setActiveView = useUIStore((s) => s.setActiveView)

  const storedAvatar = (companyData?.preferences?.avatar as string) ?? null
  const storedAvatarBg = (companyData?.preferences?.avatarBg as string) ?? '#4a7c59'

  return (
    <header className="flex items-center bg-gradient-to-r from-amber-900 via-amber-800 to-amber-900 text-white border-b-2 border-amber-700 shrink-0" style={{ height: 48 }}>
      {/* Hamburger */}
      <button
        onClick={onMenuClick}
        className="flex items-center justify-center w-11 h-full hover:bg-amber-700/30 transition-colors shrink-0"
        aria-label="Menu"
      >
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>

      {/* Company badge + name */}
      <button
        onClick={() => setActiveView('map')}
        className="flex items-center gap-2 px-2 h-full hover:bg-amber-700/30 transition-colors text-left flex-1 min-w-0"
      >
        <div
          className="w-8 h-8 rounded-full flex items-center justify-center shrink-0 overflow-hidden"
          style={{ background: storedAvatar ? undefined : storedAvatarBg }}
        >
          {storedAvatar ? (
            <img src={storedAvatar} alt="Avatar" className="w-full h-full object-cover" />
          ) : (
            <Icon name="icon_level_badge_v1" className="w-6 h-6" />
          )}
        </div>
        <div className="leading-tight min-w-0">
          <div className="font-semibold text-xs tracking-tight truncate">{companyName}</div>
          <div className="text-[9px] text-amber-300/80">Farm & Factory</div>
        </div>
      </button>

      {/* Right icons */}
      <div className="flex items-center h-full shrink-0">
        <button className="w-10 h-full flex items-center justify-center hover:bg-amber-700/30 transition-colors" title="Notifications">
          <svg className="w-4.5 h-4.5 text-amber-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
        </button>
        <button onClick={() => setActiveView('settings')} className="w-10 h-full flex items-center justify-center hover:bg-amber-700/30 transition-colors" title="Settings">
          <svg className="w-4.5 h-4.5 text-amber-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>
      </div>
    </header>
  )
}
