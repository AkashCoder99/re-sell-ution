import type { ChatConversation, ChatMessage } from '../types/chat'

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true'
const CHAT_STORAGE_KEY = 'resellution_conversations'

interface RequestOptions extends RequestInit {
  token?: string
}

async function request<TResponse>(path: string, options: RequestOptions = {}): Promise<TResponse> {
  const { token, ...fetchOptions } = options
  const headers: Record<string, string> = {
    ...((fetchOptions.headers as Record<string, string>) || {})
  }

  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  if (fetchOptions.body && typeof fetchOptions.body === 'string' && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json'
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...fetchOptions,
    headers
  })

  const data: unknown = await response.json().catch(() => ({}))

  if (!response.ok) {
    const message =
      typeof data === 'object' &&
      data !== null &&
      'error' in data &&
      typeof (data as { error?: unknown }).error === 'string'
        ? (data as { error: string }).error
        : 'Request failed'
    throw new Error(message)
  }

  return data as TResponse
}

function loadConversations(): ChatConversation[] {
  try {
    const raw = localStorage.getItem(CHAT_STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as unknown
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter((item): item is ChatConversation => typeof item === 'object' && item !== null)
      .map((conversation) => ({
        ...conversation,
        messages: Array.isArray(conversation.messages) ? conversation.messages : []
      }))
  } catch {
    return []
  }
}

function saveConversations(conversations: ChatConversation[]) {
  try {
    localStorage.setItem(CHAT_STORAGE_KEY, JSON.stringify(conversations))
  } catch {
    // ignore storage errors
  }
}

function sortConversations(conversations: ChatConversation[]): ChatConversation[] {
  return [...conversations].sort((a, b) => b.updated_at.localeCompare(a.updated_at))
}

function buildConversationId(buyerId: string, sellerId: string, listingId: string): string {
  return `${buyerId}::${sellerId}::${listingId}`
}

function withUnreadCount(conversation: ChatConversation, currentUserId?: string): ChatConversation {
  if (!currentUserId) return conversation

  const unread = conversation.messages.filter(
    (message) =>
      message.sender_id !== currentUserId &&
      (!Array.isArray(message.read_by) || !message.read_by.includes(currentUserId))
  ).length

  const lastMessage = conversation.messages[conversation.messages.length - 1]

  return {
    ...conversation,
    unread_count: unread,
    last_message_text: lastMessage?.text || ''
  }
}

export interface ListConversationsParams {
  page?: number
  limit?: number
  query?: string
  current_user_id?: string
}

export interface ListConversationsResponse {
  conversations: ChatConversation[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export async function listConversations(
  token: string,
  params: ListConversationsParams = {}
): Promise<ListConversationsResponse> {
  const page = Math.max(1, params.page ?? 1)
  const limit = Math.max(1, Math.min(25, params.limit ?? 8))
  const query = (params.query || '').trim().toLowerCase()

  if (USE_MOCK) {
    let conversations = sortConversations(loadConversations())

    if (params.current_user_id) {
      conversations = conversations.filter(
        (conversation) =>
          conversation.buyer_id === params.current_user_id || conversation.seller_id === params.current_user_id
      )
    }

    if (query) {
      conversations = conversations.filter((conversation) => {
        const lastMessage = conversation.messages[conversation.messages.length - 1]
        const haystack = [
          conversation.listing_title,
          conversation.seller_name,
          conversation.participant_name || '',
          lastMessage?.text || ''
        ]
          .join(' ')
          .toLowerCase()
        return haystack.includes(query)
      })
    }

    const total = conversations.length
    const start = (page - 1) * limit
    const items = conversations.slice(start, start + limit).map((conversation) =>
      withUnreadCount(conversation, params.current_user_id)
    )

    return {
      conversations: items,
      total,
      page,
      limit,
      total_pages: Math.max(1, Math.ceil(total / limit))
    }
  }

  const sp = new URLSearchParams()
  sp.set('page', String(page))
  sp.set('limit', String(limit))
  if (params.query) sp.set('q', params.query)

  const res = await request<Partial<ListConversationsResponse>>(`/api/v1/chat/conversations?${sp.toString()}`, {
    token
  })

  const conversations = Array.isArray(res?.conversations) ? res.conversations : []
  const total = typeof res?.total === 'number' ? res.total : conversations.length
  const totalPages =
    typeof res?.total_pages === 'number' && res.total_pages > 0
      ? res.total_pages
      : Math.max(1, Math.ceil(total / limit))

  return {
    conversations,
    total,
    page: typeof res?.page === 'number' && res.page > 0 ? res.page : page,
    limit: typeof res?.limit === 'number' && res.limit > 0 ? res.limit : limit,
    total_pages: totalPages
  }
}

export async function getConversationById(token: string, conversationId: string): Promise<ChatConversation> {
  if (USE_MOCK) {
    const conversation = loadConversations().find((item) => item.id === conversationId)
    if (!conversation) {
      throw new Error('Conversation not found.')
    }
    return conversation
  }

  const data = await request<{ conversation: ChatConversation }>(`/api/v1/chat/conversations/${conversationId}`, {
    token
  })
  return data.conversation
}

export interface StartConversationPayload {
  buyer_id: string
  seller_id: string
  seller_name: string
  listing_id: string
  listing_title: string
  listing_price: number
  listing_city: string
}

export async function getOrCreateConversation(
  token: string,
  payload: StartConversationPayload
): Promise<ChatConversation> {
  if (USE_MOCK) {
    const conversations = loadConversations()
    const id = buildConversationId(payload.buyer_id, payload.seller_id, payload.listing_id)
    const existing = conversations.find((conversation) => conversation.id === id)
    if (existing) {
      return existing
    }

    const now = new Date().toISOString()
    const conversation: ChatConversation = {
      id,
      listing_id: payload.listing_id,
      listing_title: payload.listing_title,
      listing_price: payload.listing_price,
      listing_city: payload.listing_city,
      buyer_id: payload.buyer_id,
      seller_id: payload.seller_id,
      seller_name: payload.seller_name,
      participant_name: payload.seller_name,
      created_at: now,
      updated_at: now,
      messages: [],
      unread_count: 0,
      last_message_text: ''
    }

    saveConversations([...conversations, conversation])
    return conversation
  }

  const data = await request<{ conversation: ChatConversation }>('/api/v1/chat/conversations', {
    method: 'POST',
    token,
    body: JSON.stringify(payload)
  })

  return data.conversation
}

export interface SendMessagePayload {
  sender_id: string
  text: string
}

export async function sendConversationMessage(
  token: string,
  conversationId: string,
  payload: SendMessagePayload
): Promise<ChatMessage> {
  if (USE_MOCK) {
    const conversations = loadConversations()
    const index = conversations.findIndex((conversation) => conversation.id === conversationId)
    if (index < 0) {
      throw new Error('Conversation not found.')
    }

    const now = new Date().toISOString()
    const message: ChatMessage = {
      id: `${conversationId}::${Date.now()}`,
      sender_id: payload.sender_id,
      text: payload.text,
      created_at: now,
      read_by: [payload.sender_id]
    }

    const conversation = conversations[index]
    const updatedConversation: ChatConversation = {
      ...conversation,
      updated_at: now,
      last_message_text: payload.text,
      messages: [...conversation.messages, message]
    }

    const next = [...conversations]
    next[index] = updatedConversation
    saveConversations(next)

    return message
  }

  const data = await request<{ message: ChatMessage }>(`/api/v1/chat/conversations/${conversationId}/messages`, {
    method: 'POST',
    token,
    body: JSON.stringify(payload)
  })

  return data.message
}

export async function markConversationRead(
  token: string,
  conversationId: string,
  currentUserId: string
): Promise<void> {
  if (USE_MOCK) {
    const conversations = loadConversations()
    const index = conversations.findIndex((conversation) => conversation.id === conversationId)
    if (index < 0) return

    const conversation = conversations[index]
    const updatedConversation: ChatConversation = {
      ...conversation,
      unread_count: 0,
      messages: conversation.messages.map((message) => {
        if (message.sender_id === currentUserId) return message
        const readBy = Array.isArray(message.read_by) ? message.read_by : []
        if (readBy.includes(currentUserId)) return message
        return {
          ...message,
          read_by: [...readBy, currentUserId]
        }
      })
    }

    const next = [...conversations]
    next[index] = updatedConversation
    saveConversations(next)
    return
  }

  await request(`/api/v1/chat/conversations/${conversationId}/read`, {
    method: 'PATCH',
    token
  })
}
