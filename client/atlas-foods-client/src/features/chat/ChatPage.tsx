import { useUIStore } from '@/store/ui.store'
import { MessagesView } from './MessagesView'
import { PublicChatView } from './PublicChatView'

export function ChatPage() {
  const chatTab = useUIStore(s => s.chatTab)
  const setChatTab = useUIStore(s => s.setChatTab)

  return (
    <div className="h-full flex flex-col bg-[#f5e6c8]">
      <div className="flex border-b border-amber-200/60">
        <button
          onClick={() => setChatTab('messages')}
          className={`px-4 py-2 text-xs font-bold transition-colors ${
            chatTab === 'messages'
              ? 'bg-amber-200 text-amber-900 shadow-inner'
              : 'text-amber-600 hover:bg-amber-100'
          }`}
        >
          消息
        </button>
        <button
          onClick={() => setChatTab('public')}
          className={`px-4 py-2 text-xs font-bold transition-colors ${
            chatTab === 'public'
              ? 'bg-amber-200 text-amber-900 shadow-inner'
              : 'text-amber-600 hover:bg-amber-100'
          }`}
        >
          公屏
        </button>
      </div>
      {chatTab === 'messages' ? <MessagesView /> : <PublicChatView />}
    </div>
  )
}
