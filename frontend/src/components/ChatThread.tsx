import { useMemo, useState } from 'react'
import type { ChatConversation } from '../types/chat'
import { IconBack, IconUser } from './Icons'

interface ChatThreadProps {
  conversation: ChatConversation
  currentUserId: string
  onBack: () => void
  onSendMessage: (text: string) => void
}

export default function ChatThread({
  conversation,
  currentUserId,
  onBack,
  onSendMessage
}: ChatThreadProps) {
  const [draft, setDraft] = useState('')

  const messages = useMemo(
    () => [...conversation.messages].sort((a, b) => a.created_at.localeCompare(b.created_at)),
    [conversation.messages]
  )

  return (
    <div className="chat-thread">
      <div className="chat-thread-header">
        <button type="button" className="back-link-btn" onClick={onBack}>
          <IconBack className="back-link-icon" aria-hidden />
          <span>Back</span>
        </button>
        <div className="chat-thread-title">
          <div className="chat-thread-seller">
            <IconUser className="chat-thread-seller-icon" aria-hidden />
            <div>
              <div className="chat-thread-seller-name">{conversation.seller_name}</div>
              <div className="chat-thread-seller-meta">Seller</div>
            </div>
          </div>
          <div className="chat-thread-listing">
            <span className="chat-thread-listing-title">{conversation.listing_title}</span>
            <span className="chat-thread-listing-meta">
              INR {Number(conversation.listing_price).toLocaleString()} · {conversation.listing_city}
            </span>
          </div>
        </div>
      </div>

      <div className="chat-thread-body">
        {messages.length === 0 ? (
          <div className="chat-thread-empty">
            <h3>Start the conversation</h3>
            <p>Ask about the item or negotiate with the seller.</p>
          </div>
        ) : (
          <ul className="chat-thread-messages">
            {messages.map((message) => {
              const isMine = message.sender_id === currentUserId
              return (
                <li key={message.id} className={`chat-thread-message ${isMine ? 'mine' : 'theirs'}`}>
                  <div className="chat-thread-bubble">{message.text}</div>
                  <span className="chat-thread-time">
                    {new Date(message.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </li>
              )
            })}
          </ul>
        )}
      </div>

      <form
        className="chat-thread-input"
        onSubmit={(event) => {
          event.preventDefault()
          const trimmed = draft.trim()
          if (!trimmed) return
          onSendMessage(trimmed)
          setDraft('')
        }}
      >
        <input
          type="text"
          placeholder="Type your message..."
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
        />
        <button type="submit" className="profile-edit-btn primary" disabled={!draft.trim()}>
          Send
        </button>
      </form>
    </div>
  )
}
