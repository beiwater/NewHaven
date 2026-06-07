import { useState, useEffect, useRef } from 'react'
import { useChatroomChannel, useSendMessage } from '@/api/chat.api'
import { renderMessageBody } from './ChatUtils'

const CHANNELS = [
  { id: 'general', label: '普通', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' },
  { id: 'sales', label: '销售', icon: 'M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
  { id: 'help', label: '帮助', icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
]

export function PublicChatView() {
  const [channel, setChannel] = useState('general')
  const [input, setInput] = useState('')
  const { data: messages } = useChatroomChannel(channel)
  const sendMessage = useSendMessage()
  const listRef = useRef<HTMLDivElement>(null)

  const activeChannel = CHANNELS.find(c => c.id === channel)!

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
      <div className="px-4 pt-3 pb-2">
        <div className="flex gap-1.5 p-0.5 rounded-xl bg-amber-100/60 border border-amber-200/40">
          {CHANNELS.map(ch => (
            <button
              key={ch.id}
              onClick={() => setChannel(ch.id)}
              className={`flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-lg text-[11px] font-bold transition-all ${
                channel === ch.id
                  ? 'bg-white text-amber-900 shadow-sm'
                  : 'text-amber-600 hover:text-amber-800 hover:bg-amber-50/50'
              }`}
            >
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={ch.icon} />
              </svg>
              {ch.label}
            </button>
          ))}
        </div>
      </div>

      {/* Channel description */}
      <div className="px-4 pb-2">
        <p className="text-[10px] text-amber-500 font-medium">
          {channel === 'general' && '所有人可见的普通聊天'}
          {channel === 'sales' && '交易信息、市场讨论'}
          {channel === 'help' && '提问和帮助'}
        </p>
      </div>

      {/* Messages */}
      <div ref={listRef} className="flex-1 overflow-y-auto px-4 space-y-2 pb-2">
        {(!messages || messages.length === 0) && (
          <div className="flex flex-col items-center justify-center py-12 text-amber-500">
            <svg className="w-10 h-10 mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d={activeChannel.icon} />
            </svg>
            <p className="text-[11px] font-semibold">{activeChannel.label}频道暂无消息</p>
            <p className="text-[10px] mt-1">发送第一条消息吧</p>
          </div>
        )}
        {messages?.map(msg => (
          <div key={msg.id} className="bg-white/70 rounded-xl border border-amber-200/40 px-3 py-2.5 hover:bg-white/80 transition-colors">
            <div className="flex items-center gap-1.5 mb-1">
              <div className="w-5 h-5 rounded-full bg-amber-200 flex items-center justify-center text-[9px] font-bold text-amber-800 shrink-0">
                {(msg.from || '?').charAt(0).toUpperCase()}
              </div>
              <span className="text-[10px] font-bold text-amber-800">{msg.from || 'System'}</span>
              <span className="text-[9px] text-amber-400 ml-auto">
                {msg.at ? new Date(msg.at).toLocaleTimeString() : ''}
              </span>
            </div>
            <div className="text-xs text-amber-700 leading-relaxed">{renderMessageBody(msg.body)}</div>
          </div>
        ))}
      </div>

      {/* Input */}
      <div className="border-t border-amber-200/60 p-3 flex gap-2 bg-amber-50/80">
        <input
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && handleSend()}
          placeholder={`在 ${activeChannel.label} 频道发言...`}
          className="flex-1 px-4 py-2.5 rounded-xl border border-amber-200/60 bg-white text-xs text-amber-900 placeholder-amber-300 focus:outline-none focus:ring-2 focus:ring-amber-400/40"
        />
        <button
          onClick={handleSend}
          className="px-5 py-2.5 rounded-xl bg-amber-800 text-white text-xs font-bold hover:bg-amber-900 transition-colors active:scale-95 flex items-center gap-1.5"
        >
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
          发送
        </button>
      </div>
    </div>
  )
}
