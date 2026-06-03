import { useState, useEffect, useRef, useCallback, type FormEvent } from 'react'
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

/** Minimum chat panel dimensions */
const MIN_W = 200
const MIN_H = 150

export function ChatPanel() {
  const chatOpen = useUIStore((s) => s.chatOpen)
  const setChatOpen = useUIStore((s) => s.setChatOpen)
  const [input, setInput] = useState('')
  const [showContacts, setShowContacts] = useState(false)
  const [selectedContact, setSelectedContact] = useState<{ companyId: number; company: string } | null>(null)
  const listRef = useRef<HTMLDivElement>(null)
  const panelRef = useRef<HTMLDivElement>(null)

  // Resizable state
  const [panelWidth, setPanelWidth] = useState(288) // w-72 = 288px
  const [panelHeight, setPanelHeight] = useState(400)
  const [panelPos, setPanelPos] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [isResizing, setIsResizing] = useState(false)
  const dragRef = useRef({ startX: 0, startY: 0, startW: 0, startH: 0, startPX: 0, startPY: 0 })

  const { data: messages, isLoading, error } = useMessages()
  const { data: contactsData } = useContacts()
  const sendMessage = useSendMessage()
  const markRead = useMarkRead()

  // Mark messages as read when panel opens
  useEffect(() => {
    if (chatOpen && messages && messages.length > 0) {
      markRead.mutate(messages[0].id)
    }
  }, [chatOpen])

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight
    }
  }, [messages])

  // Initialize position (bottom-right corner with some offset)
  useEffect(() => {
    if (panelPos.x === 0 && panelPos.y === 0) {
      setPanelPos({ x: window.innerWidth - panelWidth - 322, y: window.innerHeight - panelHeight - 102 })
    }
  }, [])

  // Drag and resize handlers
  const onHeaderMouseDown = useCallback((e: React.MouseEvent) => {
    if ((e.target as HTMLElement).tagName === 'BUTTON' || (e.target as HTMLElement).tagName === 'SVG') return
    e.preventDefault()
    setIsDragging(true)
    dragRef.current = { ...dragRef.current, startX: e.clientX, startY: e.clientY, startPX: panelPos.x, startPY: panelPos.y }
  }, [panelPos])

  const onResizeMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsResizing(true)
    dragRef.current = { ...dragRef.current, startX: e.clientX, startY: e.clientY, startW: panelWidth, startH: panelHeight }
  }, [panelWidth, panelHeight])

  useEffect(() => {
    if (!isDragging && !isResizing) return

    const onMove = (e: MouseEvent) => {
      if (isDragging) {
        const dx = e.clientX - dragRef.current.startX
        const dy = e.clientY - dragRef.current.startY
        setPanelPos({
          x: Math.max(0, Math.min(window.innerWidth - panelWidth, dragRef.current.startPX + dx)),
          y: Math.max(0, Math.min(window.innerHeight - panelHeight, dragRef.current.startPY + dy)),
        })
      }
      if (isResizing) {
        const dw = e.clientX - dragRef.current.startX
        const dh = e.clientY - dragRef.current.startY
        setPanelWidth(Math.max(MIN_W, dragRef.current.startW + dw))
        setPanelHeight(Math.max(MIN_H, dragRef.current.startH + dh))
      }
    }
    const onUp = () => { setIsDragging(false); setIsResizing(false) }

    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [isDragging, isResizing, panelWidth, panelHeight])

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    const text = input.trim()
    if (!text) return
    setInput('')
    sendMessage.mutate(
      { chatroom: selectedContact ? `C:${selectedContact.companyId}` : 'N', body: text },
      { onError: () => setInput(text) },
    )
  }

  const handleContactClick = (contact: { companyId: number; company: string }) => {
    setSelectedContact(contact)
    setShowContacts(false)
  }

  if (!chatOpen) return null

  const contacts = contactsData?.contacts ?? []
  const displayMessages = messages?.filter((msg) => {
    if (!selectedContact) return msg.chatroom === 'N'
    return msg.chatroom === `C:${selectedContact.companyId}`
  }) ?? []

  return (
    <div
      ref={panelRef}
      className="fixed bg-amber-50 border-2 border-amber-700/40 rounded-t-lg shadow-xl flex flex-col z-40"
      style={{
        left: panelPos.x,
        top: panelPos.y,
        width: panelWidth,
        height: panelHeight,
        cursor: isDragging ? 'grabbing' : undefined,
        userSelect: isDragging || isResizing ? 'none' : undefined,
      }}
    >
      {/* Header — draggable */}
      <div
        className="flex items-center justify-between px-3 py-2 bg-amber-800 text-white rounded-t-[5px] cursor-grab active:cursor-grabbing shrink-0"
        onMouseDown={onHeaderMouseDown}
      >
        <span className="text-xs font-semibold">
          {selectedContact ? `Chat: ${selectedContact.company}` : 'Chat'}
        </span>
        <div className="flex items-center gap-1">
          {selectedContact && (
            <button
              onClick={() => setSelectedContact(null)}
              className="text-amber-200 hover:text-white p-0.5 rounded text-[10px]"
              title="Back to public chat"
            >
              ← Global
            </button>
          )}
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
        <div className="border-b border-amber-200/60 bg-amber-100/50 px-3 py-2 space-y-1 shrink-0">
          <div className="text-[10px] uppercase tracking-wider text-amber-600 font-semibold">Contacts</div>
          {contacts.map((c) => (
            <button
              key={c.companyId}
              onClick={() => handleContactClick(c)}
              className="w-full flex items-center gap-2 text-xs text-amber-800 hover:bg-amber-200/50 rounded px-1 py-0.5 transition-colors"
            >
              <span className="w-1.5 h-1.5 rounded-full bg-amber-300" />
              <span className="font-medium">{c.company}</span>
              <span className="text-amber-500 ml-auto">Lv.{c.level}</span>
            </button>
          ))}
        </div>
      )}
      {showContacts && contacts.length === 0 && !isLoading && (
        <div className="border-b border-amber-200/60 bg-amber-100/50 px-3 py-2 shrink-0">
          <p className="text-[10px] text-amber-500 italic">No contacts yet.</p>
        </div>
      )}

      {/* Messages */}
      <div ref={listRef} className="flex-1 overflow-y-auto p-2 space-y-1.5 min-h-0">
        {isLoading && (
          <p className="text-[10px] text-amber-500 italic text-center py-4">Loading messages...</p>
        )}
        {error && (
          <p className="text-[10px] text-red-500 text-center py-4">Failed to load messages.</p>
        )}
        {!isLoading && !error && displayMessages.length === 0 && (
          <p className="text-[10px] text-amber-500 italic text-center py-4">No messages yet.</p>
        )}
        {!isLoading && !error && displayMessages.map((msg) => (
          <div key={msg.id} className="text-xs">
            <span className="font-semibold text-amber-800">
              {msg.from || (msg.chatroom === 'N' ? 'System' : msg.chatroom.replace('C:', ''))}:
            </span>{' '}
            <span className="text-amber-700">{msg.body}</span>
            <span className="text-[9px] text-amber-400 ml-1">{formatTime(msg.at)}</span>
          </div>
        ))}
        {sendMessage.isPending && (
          <p className="text-[10px] text-amber-400 italic text-right">Sending...</p>
        )}
        {sendMessage.isError && (
          <p className="text-[10px] text-red-400 text-right">Send failed.</p>
        )}
      </div>

      {/* Input */}
      <form onSubmit={handleSubmit} className="flex gap-1 p-2 border-t border-amber-200/60 shrink-0">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={selectedContact ? `Message ${selectedContact.company}...` : 'Type a message...'}
          disabled={sendMessage.isPending}
          className="flex-1 px-2 py-1 text-xs bg-white border border-amber-300 rounded text-amber-900 placeholder-amber-400 disabled:opacity-50 min-w-0"
        />
        <button
          type="submit"
          disabled={sendMessage.isPending || !input.trim()}
          className="px-3 py-1 bg-amber-700 hover:bg-amber-800 disabled:bg-amber-400 text-white text-xs font-semibold rounded transition-colors disabled:cursor-not-allowed shrink-0"
        >
          Send
        </button>
      </form>

      {/* Resize handle */}
      <div
        className="absolute bottom-0 right-0 w-4 h-4 cursor-se-resize"
        onMouseDown={onResizeMouseDown}
        style={{
          background: 'linear-gradient(135deg, transparent 50%, rgba(146, 64, 14, 0.3) 50%)',
        }}
      />
    </div>
  )
}
