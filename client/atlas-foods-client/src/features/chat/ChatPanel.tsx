import { useState, useEffect, useRef, type FormEvent } from 'react'
import { useUIStore } from '@/store/ui.store'
import { useMessages, useSendMessage, useMarkRead, useContacts } from '@/api/chat.api'

function formatTime(at: string): string {
  try {
    const d = new Date(at)
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return ''
  }
}

export function ChatPanel() {
  const chatOpen = useUIStore((s) => s.chatOpen)
  const setChatOpen = useUIStore((s) => s.setChatOpen)
  const [input, setInput] = useState('')
  const [showContacts, setShowContacts] = useState(false)
  const listRef = useRef<HTMLDivElement>(null)

  const { data: messages, isLoading, error } = useMessages()
  const { data: contactsData } = useContacts()
  const sendMessage = useSendMessage()
  const markRead = useMarkRead()

  // Mark messages as read when panel opens
  useEffect(() => {
    if (chatOpen && messages && messages.length > 0) {
      markRead.mutate(messages[0].id)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chatOpen])

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight
    }
  }, [messages])

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    const text = input.trim()
    if (!text) return

    // Optimistic: clear input immediately
    setInput('')

    sendMessage.mutate(
      { chatroom: 'N', body: text },
      {
        onError: () => {
          // Restore input on failure so user can retry
          setInput(text)
        },
      },
    )
  }

  if (!chatOpen) return null

  const contacts = contactsData?.contacts ?? []

  return (
    <div className="fixed bottom-[102px] right-[322px] w-72 max-h-96 bg-amber-50 border-2 border-amber-700/40 rounded-t-lg shadow-xl flex flex-col z-40">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-amber-800 text-white rounded-t-[5px]">
        <span className="text-xs font-semibold">Chat</span>
        <div className="flex items-center gap-1">
          <button
            onClick={() => setShowContacts((v) => !v)}
            className="text-amber-200 hover:text-white p-0.5 rounded"
            title="Contacts"
          >
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>
          <button onClick={() => setChatOpen(false)} className="text-amber-200 hover:text-white">
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      {/* Contacts panel (collapsible) */}
      {showContacts && contacts.length > 0 && (
        <div className="border-b border-amber-200/60 bg-amber-100/50 px-3 py-2 space-y-1">
          <div className="text-[10px] uppercase tracking-wider text-amber-600 font-semibold">Contacts</div>
          {contacts.map((c) => (
            <div key={c.companyId} className="flex items-center gap-2 text-xs text-amber-800">
              <span className="w-2 h-2 rounded-full bg-green-500" />
              <span className="font-medium">{c.company}</span>
              <span className="text-amber-500 ml-auto">Lv.{c.level}</span>
            </div>
          ))}
        </div>
      )}
      {showContacts && contacts.length === 0 && !isLoading && (
        <div className="border-b border-amber-200/60 bg-amber-100/50 px-3 py-2">
          <p className="text-[10px] text-amber-500 italic">No contacts yet.</p>
        </div>
      )}

      {/* Messages */}
      <div ref={listRef} className="flex-1 overflow-y-auto p-2 space-y-1.5" style={{ maxHeight: '240px' }}>
        {isLoading && (
          <p className="text-[10px] text-amber-500 italic text-center py-4">Loading messages...</p>
        )}
        {error && (
          <p className="text-[10px] text-red-500 text-center py-4">
            Failed to load messages.
          </p>
        )}
        {!isLoading && !error && messages && messages.length === 0 && (
          <p className="text-[10px] text-amber-500 italic text-center py-4">No messages yet.</p>
        )}
        {!isLoading && !error && messages?.map((msg) => (
          <div key={msg.id} className="text-xs">
            <span className="font-semibold text-amber-800">
              {msg.chatroom === 'N' ? 'System' : msg.chatroom}:
            </span>{' '}
            <span className="text-amber-700">{msg.body}</span>
            <span className="text-[9px] text-amber-400 ml-1">{formatTime(msg.at)}</span>
          </div>
        ))}
        {/* Show sending indicator */}
        {sendMessage.isPending && (
          <p className="text-[10px] text-amber-400 italic text-right">Sending...</p>
        )}
        {sendMessage.isError && (
          <p className="text-[10px] text-red-400 text-right">Send failed.</p>
        )}
      </div>

      {/* Input */}
      <form onSubmit={handleSubmit} className="flex gap-1 p-2 border-t border-amber-200/60">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type a message..."
          disabled={sendMessage.isPending}
          className="flex-1 px-2 py-1 text-xs bg-white border border-amber-300 rounded text-amber-900 placeholder-amber-400 disabled:opacity-50"
        />
        <button
          type="submit"
          disabled={sendMessage.isPending || !input.trim()}
          className="px-3 py-1 bg-amber-700 hover:bg-amber-800 disabled:bg-amber-400 text-white text-xs font-semibold rounded transition-colors disabled:cursor-not-allowed"
        >
          Send
        </button>
      </form>
    </div>
  )
}
