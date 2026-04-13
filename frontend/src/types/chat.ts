export type ChatMessage = {
  id: string
  sender_id: string
  text: string
  created_at: string
  read_by?: string[]
}

export type ChatConversation = {
  id: string
  listing_id: string
  listing_title: string
  listing_price: number
  listing_city: string
  buyer_id: string
  seller_id: string
  seller_name: string
  participant_name?: string
  created_at: string
  updated_at: string
  unread_count?: number
  last_message_text?: string
  messages: ChatMessage[]
}
