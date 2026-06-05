import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'

export interface ChatMessage {
  id: string; chatroom: string; body: string; at: string
  from?: string; fromId?: number; token?: string
}

export interface ContactEntry {
  companyId: number; company: string; playerId: string; level: number
}

export function useMessages() {
  return useQuery({
    queryKey: queryKeys.chat.messages(),
    queryFn: () => api.get<ChatMessage[]>('/api/messages/'),
    refetchInterval: (data: unknown) => (data ? 15_000 : false),
  })
}

export function useSendMessage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: { chatroom: string; body: string }) =>
      api.post<ChatMessage>('/api/v2/message/', payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.chat.messages() }),
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
    queryKey: queryKeys.chat.chatroom(),
    queryFn: () => api.get<ChatMessage[]>('/api/v2/chatroom/'),
    refetchInterval: (data: unknown) => (data ? 15_000 : false),
  })
}

export function useContacts() {
  return useQuery({
    queryKey: queryKeys.chat.contacts(),
    queryFn: () => api.get<{ contacts: ContactEntry[] }>('/api/v2/contacts/'),
    staleTime: 60_000,
  })
}
