import { useState, useEffect, useRef } from 'react'
import { useMessages, useSendMessage, useContacts } from '@/api/chat.api'
import { renderMessageBody } from './ChatUtils'
import { getCompanyId } from '@/api/client'

interface Contact {
  companyId: number
  companyName: string
  lastMessage: string
  lastTime: string
  unread: number
}

export function MessagesView() {
  const [search, setSearch] = useState('')
  const [selectedContact, setSelectedContact] = useState<{ companyId: number; companyName: string } | null>(null)
  const [input, setInput] = useState('')
  const { data: messages } = useMessages()
  const { data: contactsData } = useContacts()
  const sendMessage = useSendMessage()
  const listRef = useRef<HTMLDivElement>(null)

  const myCompanyId = Number(getCompanyId())

  // Build contact list from messages
  const messages_list = messages ?? []
  const contactMap = new Map<number, Contact>()

  // Add contacts from the contacts API
  for (const c of contactsData?.contacts ?? []) {
    if (!contactMap.has(c.companyId)) {
      contactMap.set(c.companyId, {
        companyId: c.companyId,
        companyName: c.company,
        lastMessage: '',
        lastTime: '',
        unread: 0,
      })
    }
  }

  // Add contacts from private messages
  for (const msg of messages_list) {
    if (msg.chatroom === 'N') continue
    const match = msg.chatroom.match(/^C:(\d+)$/)
    if (!match) continue
    const cid = parseInt(match[1])
    const existing = contactMap.get(cid)
    const name = msg.from || `Company-${cid}`
    if (!existing || msg.at > existing.lastTime) {
      contactMap.set(cid, {
        companyId: cid,
        companyName: name,
        lastMessage: msg.body,
        lastTime: msg.at,
        unread: existing?.unread ?? 0,
      })
    }
  }

  // Filter by search
  const contacts = Array.from(contactMap.values())
    .filter(c => c.companyName.toLowerCase().includes(search.toLowerCase()))
    .sort((a, b) => b.lastTime.localeCompare(a.lastTime))

  // Filter messages for selected contact
  const privateMessages = selectedContact
    ? messages_list.filter(msg => msg.chatroom === `C:${selectedContact.companyId}`)
    : []

  // Auto-scroll
  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight
    }
  }, [privateMessages])

  const handleSend = () => {
    if (!input.trim() || !selectedContact) return
    sendMessage.mutate({ chatroom: `C:${selectedContact.companyId}`, body: input.trim() })
    setInput('')
  }

  // Back to contacts list
  if (selectedContact) {
    return (
      <div className="flex-1 flex flex-col">
        {/* Chat header */}
        <div className="flex items-center gap-2 px-4 py-3 border-b border-amber-200/60 bg-amber-50/80">
          <button
            onClick={() => setSelectedContact(null)}
            className="flex items-center gap-1 text-[11px] font-bold text-amber-600 hover:text-amber-800 transition-colors"
          >
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            返回
          </button>
          <div className="h-4 w-px bg-amber-200/60" />
          <span className="text-sm font-bold text-amber-900">
            与 {selectedContact.companyName} 的聊天
          </span>
        </div>

        {/* Messages */}
        <div ref={listRef} className="flex-1 overflow-y-auto p-4 space-y-2">
          {privateMessages.length === 0 && (
            <p className="text-center text-[11px] text-amber-500 py-8">暂无消息，发送第一条吧</p>
          )}
          {privateMessages.map(msg => {
            const isOwn = msg.fromId !== undefined && msg.fromId === myCompanyId
            return (
              <div key={msg.id} className={`flex ${isOwn ? 'justify-end' : 'justify-start'}`}>
                <div className={`max-w-[75%] rounded-xl px-3 py-2 ${
                  isOwn
                    ? 'bg-amber-100 border border-amber-200/60'
                    : 'bg-white/70 border border-amber-200/40'
                }`}>
                  <div className={`flex items-center gap-1.5 mb-0.5 ${isOwn ? 'flex-row-reverse' : ''}`}>
                    <span className="text-[10px] font-bold text-amber-800">{msg.from || 'System'}</span>
                    <span className="text-[9px] text-amber-400">{msg.at ? new Date(msg.at).toLocaleTimeString() : ''}</span>
                  </div>
                  <div className="text-xs text-amber-700">{renderMessageBody(msg.body)}</div>
                </div>
              </div>
            )
          })}
        </div>

        {/* Input */}
        <div className="border-t border-amber-200/60 p-3 flex flex-col gap-2 bg-amber-50/80 shrink-0">
          <div className="flex gap-2">
            <input
              value={input}
              onChange={e => setInput(e.target.value.slice(0, 120))}
              onKeyDown={e => e.key === 'Enter' && handleSend()}
              placeholder="输入消息..."
              className="flex-1 px-3 py-2 rounded-xl border border-amber-200/60 bg-white text-xs text-amber-900 placeholder-amber-300 focus:outline-none focus:ring-2 focus:ring-amber-400/40"
            />
            <button
              onClick={handleSend}
              className="px-5 py-2 rounded-xl bg-amber-800 text-white text-xs font-bold hover:bg-amber-900 transition-colors active:scale-95"
            >
              发送
            </button>
          </div>
          <div className="flex justify-end">
            <span className={`text-[10px] font-medium ${
            input.length > 100 ? 'text-red-500' : input.length > 80 ? 'text-amber-500' : 'text-amber-400'
            `} }>
              {input.length} / 120
            </span>
          </div>
        </div>
      </div>
    )
  }

  // Contacts list view
  return (
    <div className="flex-1 flex flex-col">
      {/* Search */}
      <div className="px-4 pt-3 pb-2">
        <div className="relative">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            value={search}
            onChange={e => setSearch(e.target.value)}
            placeholder="搜索联系人..."
            className="w-full pl-9 pr-3 py-2 rounded-xl border border-amber-200/60 bg-white/80 text-xs text-amber-900 placeholder-amber-300 focus:outline-none focus:ring-2 focus:ring-amber-400/40"
          />
        </div>
      </div>

      {/* Contact list */}
      <div className="flex-1 overflow-y-auto">
        {contacts.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-amber-500">
            <svg className="w-10 h-10 mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <p className="text-[11px] font-semibold">暂无联系人</p>
            <p className="text-[10px] mt-1">在排行榜私聊或公屏聊天后会出现在这里</p>
          </div>
        )}
        {contacts.map(c => (
          <button
            key={c.companyId}
            onClick={() => setSelectedContact({ companyId: c.companyId, companyName: c.companyName })}
            className="w-full flex items-center gap-3 px-4 py-3 border-b border-amber-100/60 hover:bg-amber-50/80 transition-colors text-left"
          >
            <div className="w-9 h-9 rounded-full bg-amber-200 flex items-center justify-center text-sm font-bold text-amber-800 shrink-0">
              {c.companyName.charAt(0).toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-amber-900 truncate">{c.companyName}</span>
                <span className="text-[9px] text-amber-400 shrink-0 ml-2">
                  {c.lastTime ? new Date(c.lastTime).toLocaleTimeString() : ''}
                </span>
              </div>
              <p className="text-[10px] text-amber-600 truncate mt-0.5">{c.lastMessage || '暂无消息'}</p>
            </div>
          </button>
        ))}
      </div>
    </div>
  )
}
