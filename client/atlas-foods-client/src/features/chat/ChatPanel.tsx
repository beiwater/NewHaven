import { useState, type FormEvent } from 'react'
import { useUIStore } from '@/store/ui.store'

interface ChatMessage {
  id: string
  sender: string
  text: string
  time: Date
}

// Static mock messages for now
const initialMessages: ChatMessage[] = [
  { id: '1', sender: 'System', text: 'Welcome to Atlas Foods!', time: new Date() },
  { id: '2', sender: 'Market Bot', text: 'Wheat prices are up 2.1% today.', time: new Date() },
]

export function ChatPanel() {
  const chatOpen = useUIStore((s) => s.chatOpen)
  const setChatOpen = useUIStore((s) => s.setChatOpen)
  const [messages, setMessages] = useState(initialMessages)
  const [input, setInput] = useState('')

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!input.trim()) return
    const msg: ChatMessage = {
      id: crypto.randomUUID(),
      sender: 'You',
      text: input.trim(),
      time: new Date(),
    }
    setMessages((prev) => [...prev, msg])
    setInput('')
  }

  if (!chatOpen) return null

  return (
    <div className="fixed bottom-[102px] right-[322px] w-72 h-80 bg-amber-50 border-2 border-amber-700/40 rounded-t-lg shadow-xl flex flex-col z-40">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 bg-amber-800 text-white rounded-t-[5px]">
        <span className="text-xs font-semibold">Chat</span>
        <button onClick={() => setChatOpen(false)} className="text-amber-200 hover:text-white">
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-2 space-y-1.5">
        {messages.map((msg) => (
          <div key={msg.id} className="text-xs">
            <span className="font-semibold text-amber-800">{msg.sender}: </span>
            <span className="text-amber-700">{msg.text}</span>
          </div>
        ))}
      </div>

      {/* Input */}
      <form onSubmit={handleSubmit} className="flex gap-1 p-2 border-t border-amber-200/60">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type a message..."
          className="flex-1 px-2 py-1 text-xs bg-white border border-amber-300 rounded text-amber-900 placeholder-amber-400"
        />
        <button
          type="submit"
          className="px-3 py-1 bg-amber-700 hover:bg-amber-800 text-white text-xs font-semibold rounded transition-colors"
        >
          Send
        </button>
      </form>
    </div>
  )
}
