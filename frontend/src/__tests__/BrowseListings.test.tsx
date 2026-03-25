import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'

// Mock the listings API module so getCategories doesn't call real fetch
vi.mock('../api/listings', () => ({
  getCategories: vi.fn().mockResolvedValue({ categories: [] }),
  getMyListings: vi.fn().mockResolvedValue({ listings: [], total: 0, page: 1, total_pages: 1 }),
  createListing: vi.fn(),
  updateListingStatus: vi.fn(),
  deleteListing: vi.fn(),
  uploadListingImage: vi.fn()
}))

// Also stub VITE_USE_MOCK so fetchPublicListings uses mock path
vi.stubEnv('VITE_USE_MOCK', 'true')

import BrowseListings from '../components/BrowseListings'

const mockListingsResponse = {
  listings: Array.from({ length: 6 }, (_, i) => ({
    id: `browse_${i}`,
    seller_id: 'mock_seller',
    category_id: null,
    title: `Sample Item ${i + 1}`,
    description: 'A great pre-owned item.',
    condition: 'good',
    price: (i + 1) * 25,
    currency: 'USD',
    city: 'New York',
    state: null,
    status: 'active',
    view_count: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    images: []
  })),
  total: 6,
  page: 1,
  total_pages: 1
}

beforeEach(() => {
  vi.clearAllMocks()
  // Mock fetch for the listings endpoint (used when VITE_USE_MOCK env stub doesn't propagate)
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
    ok: true,
    json: async () => mockListingsResponse
  }))
})
describe('BrowseListings', () => {
  it('renders the Browse Listings heading', () => {
    render(<BrowseListings token="mock_token" userCity="New York" onBack={vi.fn()} />)
    expect(screen.getByText('Browse Listings')).toBeInTheDocument()
  })

  it('renders city filter dropdown', () => {
    render(<BrowseListings token="mock_token" onBack={vi.fn()} />)
    expect(screen.getByLabelText('Filter by city')).toBeInTheDocument()
  })

  it('renders category filter dropdown', () => {
    render(<BrowseListings token="mock_token" onBack={vi.fn()} />)
    expect(screen.getByLabelText('Filter by category')).toBeInTheDocument()
  })

  it('renders Back button and calls onBack when clicked', () => {
    const onBack = vi.fn()
    render(<BrowseListings token="mock_token" onBack={onBack} />)
    fireEvent.click(screen.getByText('← Back'))
    expect(onBack).toHaveBeenCalledTimes(1)
  })

  it('shows mock listings after loading', async () => {
    render(<BrowseListings token="mock_token" userCity="New York" onBack={vi.fn()} />)
    await waitFor(() => {
      expect(screen.getByText('Sample Item 1')).toBeInTheDocument()
    })
  })

  it('shows multiple listing cards', async () => {
    render(<BrowseListings token="mock_token" onBack={vi.fn()} />)
    await waitFor(() => {
      const cards = screen.getAllByText(/Sample Item/i)
      expect(cards.length).toBeGreaterThan(1)
    })
  })

  it('pre-selects userCity in the city dropdown', () => {
    render(<BrowseListings token="mock_token" userCity="Boston" onBack={vi.fn()} />)
    expect(screen.getByLabelText('Filter by city')).toHaveValue('Boston')
  })

  it('changes city filter selection', () => {
    render(<BrowseListings token="mock_token" onBack={vi.fn()} />)
    fireEvent.change(screen.getByLabelText('Filter by city'), { target: { value: 'Chicago' } })
    expect(screen.getByLabelText('Filter by city')).toHaveValue('Chicago')
  })
})
