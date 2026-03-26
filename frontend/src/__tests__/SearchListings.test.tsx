import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'

vi.mock('../api/listings', () => ({
  searchListings: vi.fn().mockResolvedValue({
    listings: [
      {
        id: 'l1',
        seller_id: 's1',
        category_id: null,
        title: 'Desk Lamp',
        description: 'Nice lamp',
        condition: 'good',
        price: 1200,
        currency: 'INR',
        city: 'New York',
        state: null,
        status: 'active',
        view_count: 0,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        images: []
      }
    ],
    total: 1
  }),
  getCategories: vi.fn().mockResolvedValue({ categories: [] })
}))

import SearchListings from '../components/SearchListings'

beforeEach(() => {
  localStorage.clear()
})

describe('SearchListings', () => {
  it('renders search input and button', () => {
    render(<SearchListings token="t" userCity="New York" onBack={vi.fn()} />)
    expect(screen.getByPlaceholderText('Search in New York')).toBeInTheDocument()
    expect(screen.getByText('Search')).toBeInTheDocument()
  })

  it('shows recent searches when focused and empty', () => {
    localStorage.setItem('resellution_recent_searches', JSON.stringify(['lamp']))
    render(<SearchListings token="t" userCity="New York" onBack={vi.fn()} />)
    const input = screen.getByPlaceholderText('Search in New York')
    fireEvent.focus(input)
    expect(screen.getByText('Recent searches')).toBeInTheDocument()
    expect(screen.getByText('lamp')).toBeInTheDocument()
  })

  it('runs search and shows a result card', async () => {
    render(<SearchListings token="t" userCity="New York" onBack={vi.fn()} />)
    fireEvent.change(screen.getByPlaceholderText('Search in New York'), { target: { value: 'lamp' } })
    fireEvent.click(screen.getByText('Search'))
    await waitFor(() => {
      expect(screen.getByText('Desk Lamp')).toBeInTheDocument()
    })
  })
})
