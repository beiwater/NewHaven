import { useState, useEffect, useRef } from 'react'
import { useChatroomChannel, useSendMessage } from '@/api/chat.api'
import { renderMessageBody } from './ChatUtils'

export function PublicChatView() {
  const [channel, setChannel] = useState('general')
  const [input, setInput] = useState('')
  const { data: messages } = useChatroomChannel(channel)
  const sendMessage = useSendMessage()
  const listRef = useRef<HTMLDivElement>(null)

  const channels = [
    { id: 'general', label: '普通' },
    { id: 'sales', label: '销售' },
    { id: 'help', label: '帮助' },
  ]

  // Auto scroll to bottom
  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight
    }
  }, [messages])

  const handleSend = () => {
    if (!input.trim()) return
    sendMessage.mutate({ chatroom: channel, body: input.trim() })
    setInput('')
  }

  return (
    <div className="flex-1 flex flex-col">
      {/* Channel tabs */}
      <div className="flex border-b border-amber-200/60 bg-amber-50/50">
        {channels.map(ch => (
          <button
            key={ch.id}
            onClick={() => setChannel(ch.id)}
            className={`px-4 py-2 text-xs font-bold ${
              channel === ch.id
                ? 'bg-amber-200 text-amber-900'
                : 'text-amber-600 hover:bg-amber-100'
            }`}
          >
            {ch.label}
          </button>
        ))}
      </div>

      {/* Messages */}
      <div ref={listRef} className="flex-1 overflow-y-auto p-4 space-y-2">
        {messages?.map(msg => (
          <div key={msg.id} className="text-xs">
            <span className="font-bold text-amber-800">{msg.from || 'System'}:</span>
            {' '}
            {renderMessageBody(msg.body)}
            <span className="text-[9px] text-amber-400 ml-1">
              {msg.at ? new Date(msg.at).toLocaleTimeString() : ''}
            </span>
          </div>
        ))}
      </div>

      {/* Input */}
      <div className="border-t border-amber-200/60 p-3 flex gap-2 bg-amber-50/80">
        <input
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleSend()}
          placeholder="输入消息..."
          className="flex-1 px-3 py-2 rounded-lg border border-amber-200/60 bg-white text-xs"
        />
        <button
          onClick={handleSend}
          className="px-4 py-2 rounded-lg bg-amber-800 text-white text-xs font-bold hover:bg-amber-900"
        >
          发送
        </button>
      </div>
    </div>
  )
}
