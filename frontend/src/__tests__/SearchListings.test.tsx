import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'

const { searchListingsMock, getCategoriesMock } = vi.hoisted(() => ({
  searchListingsMock: vi.fn(),
  getCategoriesMock: vi.fn()
}))

vi.mock('../api/listings', () => ({
  searchListings: searchListingsMock,
  getCategories: getCategoriesMock
}))

import SearchListings from '../components/SearchListings'

function ensureLocalStorage() {
  const candidate = globalThis.localStorage as
    | {
        getItem?: (key: string) => string | null
        setItem?: (key: string, value: string) => void
        removeItem?: (key: string) => void
        clear?: () => void
      }
    | undefined

  const hasAPI =
    candidate &&
    typeof candidate.getItem === 'function' &&
    typeof candidate.setItem === 'function' &&
    typeof candidate.removeItem === 'function' &&
    typeof candidate.clear === 'function'

  if (hasAPI) return

  const store = new Map<string, string>()
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    value: {
      getItem: (key: string) => (store.has(key) ? store.get(key)! : null),
      setItem: (key: string, value: string) => {
        store.set(key, value)
      },
      removeItem: (key: string) => {
        store.delete(key)
      },
      clear: () => {
        store.clear()
      }
    }
  })
}

beforeEach(() => {
  ensureLocalStorage()
  localStorage.clear()
  vi.clearAllMocks()
  searchListingsMock.mockResolvedValue({
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
  })
  getCategoriesMock.mockResolvedValue({ categories: [] })
})

describe('SearchListings', () => {
  it('renders search input and button', async () => {
    render(<SearchListings token="t" userCity="New York" onBack={vi.fn()} />)
    await waitFor(() => {
      expect(getCategoriesMock).toHaveBeenCalled()
    })
    expect(screen.getByPlaceholderText('Search in New York')).toBeInTheDocument()
    expect(screen.getByText('Search')).toBeInTheDocument()
  })

  it('shows recent searches when focused and empty', async () => {
    localStorage.setItem('resellution_recent_searches', JSON.stringify(['lamp']))
    render(<SearchListings token="t" userCity="New York" onBack={vi.fn()} />)
    await waitFor(() => {
      expect(getCategoriesMock).toHaveBeenCalled()
    })
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
