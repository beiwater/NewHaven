import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from './client'

// --- Types matching backend model.Message ---
export interface ChatMessage {
  id: string
  chatroom: string
  body: string
  at: string
  from?: string
  fromId?: number
  token?: string
}

export interface ContactEntry {
  companyId: number
  company: string
  playerId: string
  level: number
}

export interface ContactsResponse {
  chatrooms: unknown[]
  contacts: ContactEntry[]
  unreadMessagesOtherRealms: number
  invisible: boolean
  ignoringCompanies: Record<string, unknown>
  companiesChatBlockingUs: Record<string, unknown>
}

export interface SendMessagePayload {
  chatroom: string
  body: string
  token?: string
}

// --- Hooks ---

export function useMessages() {
  return useQuery({
    queryKey: ['chat', 'messages'],
    queryFn: () => api.get<ChatMessage[]>('/api/messages/'),
    refetchInterval: (data) => (data ? 15_000 : false),
    retry: 2,
    retryDelay: 5000,
  })
}

export function useSendMessage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: SendMessagePayload) =>
      api.post<ChatMessage>('/api/v2/message/', payload),
    onSuccess: (_, variables) => {
      qc.invalidateQueries({ queryKey: ['chat', 'messages'] })
      qc.invalidateQueries({ queryKey: ['chat', 'chatroom'] })
      if (variables.chatroom.startsWith('C:')) {
        qc.invalidateQueries({ queryKey: ['chat', 'private'] })
      }
    },
  })
}

export function useMarkRead() {
  return useMutation({
    mutationFn: (messageId: string) =>
      api.get<{ ok: boolean }>(`/api/v2/message/${messageId}/read/`),
  })
}

export function useChatroom() {
  return useQuery({
    queryKey: ['chat', 'chatroom'],
    queryFn: () => api.get<ChatMessage[]>('/api/v2/chatroom/'),
    refetchInterval: (data) => (data ? 15_000 : false),
    retry: 2,
    retryDelay: 5000,
  })
}

export function useChatroomChannel(channel: string) {
  return useQuery({
    queryKey: ['chat', 'chatroom', channel],
    queryFn: () => api.get<ChatMessage[]>(`/api/v2/chatroom/?channel=${channel}`),
    refetchInterval: (data) => (data ? 15_000 : false),
    retry: 2,
    retryDelay: 5000,
  })
}

export function usePrivateMessages(companyId: number | null) {
  return useQuery({
    queryKey: ['chat', 'private', companyId],
    queryFn: () => api.get<ChatMessage[]>(`/api/messages/?chatroom=C:${companyId}`),
    enabled: companyId !== null,
    refetchInterval: (data) => (data ? 15_000 : false),
    retry: 2,
    retryDelay: 5000,
  })
}

export function useContacts() {
  return useQuery({
    queryKey: ['chat', 'contacts'],
    queryFn: () => api.get<ContactsResponse>('/api/v2/contacts/'),
    staleTime: 60_000,
  })
}

// --- Room-based chat types ---
export interface ChatRoom {
  id: string
  participant1: number
  participant2: number
  last_message_at?: string
  created_at: string
}
export interface CreateRoomResponse {
  room: ChatRoom
  messages: ChatMessage[]
}
export interface RoomListResponse {
  rooms: ChatRoom[]
}
export interface RoomMessagesResponse {
  messages: ChatMessage[]
}
export function useChatRooms() {
  return useQuery({
    queryKey: ['chat', 'rooms'],
    queryFn: () => api.get<RoomListResponse>('/api/v2/chat/rooms/'),
    refetchInterval: 15_000,
  })
}
export function useCreateRoom() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (companyId: number) =>
      api.post<CreateRoomResponse>('/api/v2/chat/room/', { companyId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['chat', 'rooms'] })
    },
  })
}
export function useRoomMessages(roomId: string | null) {
  return useQuery({
    queryKey: ['chat', 'room', roomId, 'messages'],
    queryFn: () => api.get<RoomMessagesResponse>(`/api/v2/chat/room/${roomId}/messages/`),
    enabled: !!roomId,
    refetchInterval: (data) => (data ? 15_000 : false),
  })
}
export function useSendRoomMessage(roomId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: string) =>
      api.post<ChatMessage>('/api/v2/chat/room/send/', { roomId, body }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['chat', 'room', roomId, 'messages'] })
    },
  })
}
