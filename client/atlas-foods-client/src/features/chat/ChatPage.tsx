import { useUIStore } from '@/store/ui.store'
import { MessagesView } from './MessagesView'
import { PublicChatView } from './PublicChatView'

export function ChatPage() {
  const chatTab = useUIStore(s => s.chatTab)
  const setChatTab = useUIStore(s => s.setChatTab)

  return (
    <div className="h-full flex flex-col">
      <div className="mx-auto w-full max-w-3xl p-4 md:p-6 flex flex-col flex-1 min-h-0">
        {/* Header */}
        <div className="mb-5 flex items-center gap-3 shrink-0">
          <svg className="h-7 w-7 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          <div>
            <p className="text-[10px] font-bold uppercase tracking-[0.24em] text-amber-700/70">
              Social
            </p>
            <h2 className="text-2xl font-black text-amber-950">聊天</h2>
          </div>
        </div>

        {/* Tab pills */}
        <div className="flex gap-1.5 p-0.5 rounded-xl bg-amber-100/60 border border-amber-200/40 mb-4 shrink-0">
          <button
            onClick={() => setChatTab('messages')}
            className={`flex-1 flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
              chatTab === 'messages'
                ? 'bg-white text-amber-900 shadow-sm'
                : 'text-amber-600 hover:text-amber-800 hover:bg-amber-50/50'
            }`}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            消息
          </button>
          <button
            onClick={() => setChatTab('public')}
            className={`flex-1 flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg text-xs font-bold transition-all ${
              chatTab === 'public'
                ? 'bg-white text-amber-900 shadow-sm'
                : 'text-amber-600 hover:text-amber-800 hover:bg-amber-50/50'
            }`}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
            </svg>
            公屏
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 bg-white/50 rounded-2xl border border-amber-200/50 shadow-sm overflow-hidden flex flex-col min-h-0">
          {chatTab === 'messages' ? <MessagesView /> : <PublicChatView />}
        </div>
      </div>
    </div>
  )
}
