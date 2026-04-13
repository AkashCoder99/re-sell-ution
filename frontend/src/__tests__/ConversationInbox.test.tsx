import { beforeEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'

const listConversationsMock = vi.fn()

vi.mock('../api/chat', () => ({
  listConversations: (...args: unknown[]) => listConversationsMock(...args)
}))

import ConversationInbox from '../components/ConversationInbox'

function buildConversation(id: string, updates: Partial<Record<string, unknown>> = {}) {
  return {
    id,
    listing_id: `listing-${id}`,
    listing_title: `Listing ${id}`,
    listing_price: 100,
    listing_city: 'New York',
    buyer_id: 'buyer-1',
    seller_id: 'seller-1',
    seller_name: 'Seller Name',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    unread_count: 0,
    last_message_text: 'Hello there',
    messages: [],
    ...updates
  }
}

describe('ConversationInbox', () => {
  beforeEach(() => {
    listConversationsMock.mockReset()
  })

  it('renders conversations with unread badge and opens selected thread', async () => {
    listConversationsMock.mockResolvedValue({
      conversations: [buildConversation('c1', { unread_count: 3, participant_name: 'Alex' })],
      total: 1,
      page: 1,
      limit: 6,
      total_pages: 1
    })

    const onOpenConversation = vi.fn()

    render(
      <ConversationInbox
        token="token"
        currentUserId="buyer-1"
        onBack={vi.fn()}
        onOpenConversation={onOpenConversation}
      />
    )

    await waitFor(() => {
      expect(screen.getByText('Listing c1')).toBeInTheDocument()
    })

    expect(screen.getByText('3')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /alex/i }))
    expect(onOpenConversation).toHaveBeenCalledWith('c1')
  })

  it('supports pagination controls and fetches next page', async () => {
    listConversationsMock
      .mockResolvedValueOnce({
        conversations: [buildConversation('c1')],
        total: 2,
        page: 1,
        limit: 6,
        total_pages: 2
      })
      .mockResolvedValueOnce({
        conversations: [buildConversation('c2')],
        total: 2,
        page: 2,
        limit: 6,
        total_pages: 2
      })

    render(
      <ConversationInbox
        token="token"
        currentUserId="buyer-1"
        onBack={vi.fn()}
        onOpenConversation={vi.fn()}
      />
    )

    await waitFor(() => {
      expect(screen.getByText('Listing c1')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByRole('button', { name: 'Next' }))

    await waitFor(() => {
      expect(screen.getByText('Listing c2')).toBeInTheDocument()
    })

    expect(listConversationsMock).toHaveBeenLastCalledWith(
      'token',
      expect.objectContaining({ page: 2, current_user_id: 'buyer-1' })
    )
  })
})
