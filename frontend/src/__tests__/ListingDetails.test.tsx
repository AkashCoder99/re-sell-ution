import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import ListingDetails from '../components/ListingDetails'
import type { Listing } from '../types/listing'

const baseListing: Listing = {
  id: 'l1',
  seller_id: 's1',
  category_id: null,
  title: 'Vintage Desk Lamp',
  description: 'Great lamp',
  condition: 'good',
  price: 1500,
  currency: 'INR',
  city: 'New York',
  state: null,
  status: 'active',
  view_count: 0,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  images: [
    { id: 'i1', listing_id: 'l1', image_url: 'https://placehold.co/400x300?text=1', position: 0 },
    { id: 'i2', listing_id: 'l1', image_url: 'https://placehold.co/400x300?text=2', position: 1 }
  ]
}

describe('ListingDetails', () => {
  it('renders title and price', () => {
    render(<ListingDetails listing={baseListing} onClose={() => {}} />)
    expect(screen.getByText('Vintage Desk Lamp')).toBeInTheDocument()
    expect(screen.getByText('INR 1,500')).toBeInTheDocument()
  })

  it('navigates images with next button', () => {
    render(<ListingDetails listing={baseListing} onClose={() => {}} />)
    const next = screen.getByText('>')
    fireEvent.click(next)
    expect(screen.getAllByAltText('Vintage Desk Lamp').length).toBeGreaterThan(0)
  })

  it('shows unavailable message for deleted listing', () => {
    render(
      <ListingDetails listing={{ ...baseListing, status: 'deleted' }} onClose={() => {}} />
    )
    expect(screen.getByText('This listing is no longer available.')).toBeInTheDocument()
  })

  it('calls favorite toggle handler with the next state', async () => {
    const onToggleFavorite = vi.fn().mockResolvedValue(undefined)
    render(
      <ListingDetails
        listing={baseListing}
        onClose={() => {}}
        isFavorite={false}
        onToggleFavorite={onToggleFavorite}
      />
    )

    fireEvent.click(screen.getByText('Favorite'))

    await waitFor(() => {
      expect(onToggleFavorite).toHaveBeenCalledWith(baseListing, true)
    })
  })

  it('reports a listing and shows success feedback', async () => {
    const onReport = vi.fn().mockResolvedValue(undefined)
    render(
      <ListingDetails
        listing={baseListing}
        onClose={() => {}}
        onReport={onReport}
      />
    )

    fireEvent.click(screen.getByText('Report'))

    await waitFor(() => {
      expect(onReport).toHaveBeenCalledWith(baseListing)
    })
    expect(await screen.findByText('Listing reported.')).toBeInTheDocument()
  })
})
