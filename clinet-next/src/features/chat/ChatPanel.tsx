import { useState } from 'react'
import { useMessages, useSendMessage, useChatroom } from '@/api/hooks/chat.hooks'

export function ChatPanel() {
  const [body, setBody] = useState('')
  const { data: messages } = useMessages()
  const { data: chatroom } = useChatroom()
  const sendMessage = useSendMessage()

  const allMessages = chatroom ?? messages ?? []

  const handleSend = () => {
    if (!body.trim()) return
    sendMessage.mutate({ chatroom: 'general', body: body.trim() })
    setBody('')
  }

  return (
    <div className="flex flex-col h-full p-4">
      <h2 className="text-lg font-bold text-amber-900 mb-3">Chat</h2>
      <div className="flex-1 overflow-y-auto space-y-2 mb-3">
        {allMessages.length === 0 && (
          <div className="text-xs text-amber-400 italic text-center py-8">No messages yet</div>
        )}
        {allMessages.map((m, i) => (
          <div key={m.id ?? i} className="p-2 bg-white/60 rounded-lg border border-amber-200/40 text-xs">
            <div className="flex justify-between text-[10px] text-amber-600 mb-1">
              <span className="font-semibold">{m.from ?? 'System'}</span>
              <span>{new Date(m.at).toLocaleTimeString()}</span>
            </div>
            <div className="text-amber-900">{m.body}</div>
          </div>
        ))}
      </div>
      <div className="flex gap-2">
        <input
          value={body}
          onChange={(e) => setBody(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          placeholder="Type a message..."
          className="flex-1 px-3 py-1.5 text-xs border border-amber-200 rounded-md bg-white focus:outline-none focus:ring-1 focus:ring-amber-700"
        />
        <button
          onClick={handleSend}
          disabled={sendMessage.isPending}
          className="px-3 py-1.5 bg-amber-600 text-white text-xs font-semibold rounded-md hover:bg-amber-700 disabled:opacity-50"
        >
          Send
        </button>
      </div>
    </div>
  )
}
