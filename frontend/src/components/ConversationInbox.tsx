import { useEffect, useMemo, useState } from 'react'
import { listConversations } from '../api/chat'
import type { ChatConversation } from '../types/chat'
import { IconBack, IconSearch, IconUser } from './Icons'

interface ConversationInboxProps {
  token: string
  currentUserId: string
  onBack: () => void
  onOpenConversation: (conversationId: string) => void | Promise<void>
  refreshKey?: number
}

const PAGE_SIZE = 6

function formatUpdatedAt(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return 'Unknown activity'

  const diffMinutes = Math.floor((Date.now() - date.getTime()) / 60000)
  if (diffMinutes < 1) return 'Just now'
  if (diffMinutes < 60) return `${diffMinutes}m ago`
  const diffHours = Math.floor(diffMinutes / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  const diffDays = Math.floor(diffHours / 24)
  return `${diffDays}d ago`
}

export default function ConversationInbox({
  token,
  currentUserId,
  onBack,
  onOpenConversation,
  refreshKey = 0
}: ConversationInboxProps) {
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [conversations, setConversations] = useState<ChatConversation[]>([])
  const [totalPages, setTotalPages] = useState(1)

  useEffect(() => {
    let cancelled = false

    async function loadInbox() {
      setLoading(true)
      setError('')
      try {
        const data = await listConversations(token, {
          page,
          limit: PAGE_SIZE,
          query,
          current_user_id: currentUserId
        })

        if (cancelled) return

        setConversations(data.conversations)
        setTotalPages(data.total_pages)
      } catch (err: unknown) {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'Unable to load your inbox.')
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadInbox()

    return () => {
      cancelled = true
    }
  }, [token, currentUserId, page, query, refreshKey])

  const hasConversations = conversations.length > 0
  const pageLabel = useMemo(() => `Page ${page} of ${Math.max(1, totalPages)}`, [page, totalPages])

  return (
    <div className="conversation-inbox">
      <div className="conversation-inbox-header">
        <button type="button" className="back-link-btn" onClick={onBack}>
          <IconBack className="back-link-icon" aria-hidden />
          <span>Back</span>
        </button>
        <h2>Inbox</h2>
      </div>

      <div className="conversation-inbox-search profile-search-input">
        <IconSearch className="profile-search-icon" aria-hidden />
        <input
          type="text"
          placeholder="Search conversations"
          value={query}
          onChange={(event) => {
            setQuery(event.target.value)
            setPage(1)
          }}
        />
      </div>

      {error && <p className="conversation-inbox-error">{error}</p>}

      {loading ? (
        <p className="conversation-inbox-loading">Loading conversations…</p>
      ) : hasConversations ? (
        <ul className="conversation-inbox-list">
          {conversations.map((conversation) => {
            const unreadCount = conversation.unread_count ?? 0
            const preview = conversation.last_message_text || 'No messages yet. Start the conversation.'
            const participant =
              conversation.participant_name ||
              (conversation.seller_id === currentUserId ? conversation.buyer_id : conversation.seller_name)

            return (
              <li key={conversation.id}>
                <button
                  type="button"
                  className="conversation-inbox-item"
                  onClick={() => {
                    void onOpenConversation(conversation.id)
                  }}
                >
                  <div className="conversation-inbox-item-main">
                    <div className="conversation-inbox-item-top">
                      <div className="conversation-inbox-participant">
                        <IconUser className="conversation-inbox-user-icon" aria-hidden />
                        <span>{participant}</span>
                      </div>
                      <span className="conversation-inbox-time">{formatUpdatedAt(conversation.updated_at)}</span>
                    </div>
                    <div className="conversation-inbox-listing">{conversation.listing_title}</div>
                    <div className="conversation-inbox-preview">{preview}</div>
                  </div>
                  {unreadCount > 0 && <span className="conversation-inbox-unread">{unreadCount}</span>}
                </button>
              </li>
            )
          })}
        </ul>
      ) : (
        <div className="conversation-inbox-empty">
          <h3>No conversations yet</h3>
          <p>Start chatting from a listing and your conversations will appear here.</p>
        </div>
      )}

      <div className="conversation-inbox-pagination">
        <button
          type="button"
          className="profile-edit-btn secondary"
          onClick={() => setPage((prev) => Math.max(1, prev - 1))}
          disabled={loading || page <= 1}
        >
          Previous
        </button>
        <span>{pageLabel}</span>
        <button
          type="button"
          className="profile-edit-btn secondary"
          onClick={() => setPage((prev) => Math.min(totalPages, prev + 1))}
          disabled={loading || page >= totalPages}
        >
          Next
        </button>
      </div>
    </div>
  )
}
