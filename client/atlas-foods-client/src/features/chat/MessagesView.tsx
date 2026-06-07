import { useState, useEffect, useRef } from 'react'
import { useChatRooms, useRoomMessages, useSendRoomMessage, useContacts } from '@/api/chat.api'
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
  const [selectedRoomId, setSelectedRoomId] = useState<string | null>(null)
  const [selectedPartnerName, setSelectedPartnerName] = useState('')
  const [input, setInput] = useState('')
  const [chatError, setChatError] = useState('')
  const myCompanyId = Number(getCompanyId())

  const { data: roomsData } = useChatRooms()
  const { data: contactsData } = useContacts()
  const { data: roomMessages } = useRoomMessages(selectedRoomId)
  const sendMessage = useSendRoomMessage(selectedRoomId ?? '')
  const listRef = useRef<HTMLDivElement>(null)

  const rooms = roomsData?.rooms ?? []
  const contacts = contactsData?.contacts ?? []
  const messages = roomMessages?.messages ?? []

  // Build contact list from rooms (existing conversations)
  const chatEntries: (Contact & { roomId: string })[] = rooms
    .map(room => {
      const otherId = room.participant1 === myCompanyId ? room.participant2 : room.participant1
      const contact = contacts.find(c => c.companyId === otherId)
      return {
        roomId: room.id,
        companyId: otherId,
        companyName: contact?.company ?? `Company-${otherId}`,
		lastMessage: room.last_message ?? '',
        lastTime: room.last_message_at ?? '',
        unread: 0,
      }
    })
    .sort((a, b) => b.lastTime.localeCompare(a.lastTime))

  // Search results: when 3+ chars, search ALL companies from contacts
  const chatRoomIds = new Set(rooms.map(r => r.participant1 === myCompanyId ? r.participant2 : r.participant1))
  const searchResults = search.length >= 3
    ? contacts
        .filter(c => !chatRoomIds.has(c.companyId) && c.company.toLowerCase().includes(search.toLowerCase()))
        .slice(0, 10)
    : []

  // Auto-scroll
  useEffect(() => {
    if (listRef.current) {
      listRef.current.scrollTop = listRef.current.scrollHeight
    }
  }, [messages])

  // Mark messages as read when room opens
  useEffect(() => {
    if (selectedRoomId && messages.length > 0) {
      const lastId = messages[messages.length - 1].id
      fetch(`/api/v2/chat/room/${selectedRoomId}/read/`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${localStorage.getItem('atlas_auth_token')}` },
        body: JSON.stringify({ lastMessageId: lastId }),
      })
    }
  }, [selectedRoomId, messages])
  const handleSelectContact = async (contact: Contact) => {
    setChatError('')
    setSelectedPartnerName(contact.companyName)
    const existing = rooms.find(r =>
      (r.participant1 === myCompanyId && r.participant2 === contact.companyId) ||
      (r.participant1 === contact.companyId && r.participant2 === myCompanyId)
    )
    if (existing) {
      setSelectedRoomId(existing.id)
      return
    }
    // Create room via direct fetch (bypass mutation to avoid state issues)
    try {
      const res = await fetch('/api/v2/chat/room/', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${localStorage.getItem('atlas_auth_token')}` },
        body: JSON.stringify({ companyId: contact.companyId }),
      })
      if (!res.ok) {
        setChatError(`创建聊天室失败 (${res.status})`)
        return
      }
      const data = await res.json()
      if (data?.room?.id) {
        setSelectedRoomId(data.room.id)
      }
    } catch (err) {
      setChatError('网络错误')
    }
  }
  const handleSend = () => {
    if (!input.trim() || !selectedRoomId) return
    sendMessage.mutate(input.trim())
    setInput('')
  }

  // Back to contacts list
  if (selectedRoomId) {
    return (
      <div className="flex-1 flex flex-col min-h-0">
        {/* Chat header */}
        <div className="flex items-center gap-2 px-4 py-3 border-b border-amber-200/60 bg-amber-50/80">
          <button
            onClick={() => { setSelectedRoomId(null); setSelectedPartnerName('') }}
            className="flex items-center gap-1 text-[11px] font-bold text-amber-600 hover:text-amber-800 transition-colors"
          >
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            返回
          </button>
          <div className="h-4 w-px bg-amber-200/60" />
          <span className="text-sm font-bold text-amber-900">
            与 {selectedPartnerName} 的聊天
          </span>
        </div>

        {/* Messages */}
        <div ref={listRef} className="flex-1 overflow-y-auto min-h-0 p-4 space-y-2">
          {messages.length === 0 && (
            <p className="text-center text-[11px] text-amber-500 py-8">暂无消息，发送第一条吧</p>
          )}
          {messages.map(msg => {
            const isOwn = msg.sender_id === myCompanyId
            return (
              <div key={msg.id} className={`flex ${isOwn ? 'justify-end' : 'justify-start'}`}>
                <div className={`max-w-[75%] rounded-xl px-3 py-2 ${
                  isOwn
                    ? 'bg-amber-100 border border-amber-200/60'
                    : 'bg-white/70 border border-amber-200/40'
                }`}>
                  <div className={`flex items-center gap-1.5 mb-0.5 ${isOwn ? 'flex-row-reverse' : ''}`}>
                    <span className="text-[10px] font-bold text-amber-800">{msg.sender_name || 'System'}</span>
                    <span className="text-[9px] text-amber-400">{msg.created_at ? new Date(msg.created_at).toLocaleTimeString() : ''}</span>
                  </div>
                  <div className="text-xs text-amber-700">{renderMessageBody(msg.content)}</div>
                  {isOwn && (msg as any).read && (
                    <span className="text-[9px] text-amber-400 text-right block mt-0.5">已读</span>
                  )}
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
            }`}>
              {input.length} / 120
            </span>
          </div>
        </div>
      </div>
    )
  }

  // Contacts list view
  return (
    <div className="flex-1 flex flex-col min-h-0">
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

      {/* Error message */}
      {chatError && (
        <div className="px-4 py-2 text-[10px] text-red-500 font-medium bg-red-50 border-b border-red-200/60">
          {chatError}
        </div>
      )}

      {/* Contact list */}
      <div className="flex-1 overflow-y-auto">
        {/* Existing conversations */}
        {chatEntries.length > 0 && chatEntries.map(c => (
          <button
            key={c.roomId}
            onClick={() => handleSelectContact(c)}
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

        {/* Search results (companies not yet chatted with) */}
        {searchResults.length > 0 && (
          <>
            <div className="px-4 py-2 text-[9px] font-bold uppercase tracking-wider text-amber-500 bg-amber-50/50">
              搜索到 {search.length >= 3 ? `${contacts.filter(c => c.company.toLowerCase().includes(search.toLowerCase())).length} 个结果` : ''}
            </div>
            {searchResults.map(c => (
              <button
                key={`sr-${c.companyId}`}
                onClick={() => handleSelectContact({ companyId: c.companyId, companyName: c.company, lastMessage: '', lastTime: '', unread: 0 })}
                className="w-full flex items-center gap-3 px-4 py-3 border-b border-amber-100/60 hover:bg-amber-50/80 transition-colors text-left"
              >
                <div className="w-9 h-9 rounded-full bg-amber-300/60 flex items-center justify-center text-sm font-bold text-amber-700 shrink-0">
                  {c.company.charAt(0).toUpperCase()}
                </div>
                <div className="flex-1 min-w-0">
                  <span className="text-xs font-bold text-amber-900 truncate">{c.company}</span>
                  <p className="text-[10px] text-amber-500 mt-0.5">开始聊天</p>
                </div>
              </button>
            ))}
          </>
        )}

        {/* Empty state */}
        {chatEntries.length === 0 && searchResults.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 text-amber-500">
            <svg className="w-10 h-10 mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <p className="text-[11px] font-semibold">暂无联系人</p>
            <p className="text-[10px] mt-1">搜索公司名称开始聊天</p>
          </div>
        )}
      </div>
    </div>
  )
}
