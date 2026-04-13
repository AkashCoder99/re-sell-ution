import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import type { Category, Listing } from '../types/listing'
import { LISTING_CONDITIONS } from '../types/listing'
import { getCategories, searchListings } from '../api/listings'
import { IconBack, IconSearch } from './Icons'
import ListingDetails from './ListingDetails'

interface SearchListingsProps {
  token: string
  userCity: string
  onBack: () => void
  onStartChat?: (listing: Listing) => Promise<void>
}

const RECENT_KEY = 'resellution_recent_searches'
const PREFILL_KEY = 'resellution_search_prefill'
const FILTERS_KEY = 'resellution_search_filters'
const MAX_RECENT = 6
const PAGE_SIZE = 8

type SortOption = 'newest' | 'price_low' | 'price_high'

type Filters = {
  minPrice: string
  maxPrice: string
  categoryId: string
  condition: string
  city: string
  sort: SortOption
}

const defaultFilters: Filters = {
  minPrice: '',
  maxPrice: '',
  categoryId: '',
  condition: '',
  city: '',
  sort: 'newest'
}

function loadRecent(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as unknown
    if (!Array.isArray(parsed)) return []
    return parsed.filter((v) => typeof v === 'string').slice(0, MAX_RECENT)
  } catch {
    return []
  }
}

function saveRecent(list: string[]) {
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(list.slice(0, MAX_RECENT)))
  } catch {
    // ignore localStorage failures
  }
}

export default function SearchListings({ token, userCity, onBack, onStartChat }: SearchListingsProps) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<Listing[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [hasSearched, setHasSearched] = useState(false)
  const [recent, setRecent] = useState<string[]>([])
  const [page, setPage] = useState(1)
  const [lastQuery, setLastQuery] = useState('')
  const [showRecent, setShowRecent] = useState(false)
  const [selected, setSelected] = useState<Listing | null>(null)
  const [filters, setFilters] = useState<Filters>(defaultFilters)
  const [draftFilters, setDraftFilters] = useState<Filters>(defaultFilters)
  const [showFilters, setShowFilters] = useState(false)
  const [filtering, setFiltering] = useState(false)

  useEffect(() => {
    setRecent(loadRecent())
    getCategories(token)
      .then((res) => setCategories(Array.isArray(res.categories) ? res.categories : []))
      .catch(() => setCategories([]))
    try {
      const prefill = localStorage.getItem(PREFILL_KEY) || ''
      if (prefill) {
        setQuery(prefill)
        localStorage.removeItem(PREFILL_KEY)
        void runSearch(prefill)
      }
      const savedFilters = localStorage.getItem(FILTERS_KEY)
      if (savedFilters) {
        const parsed = JSON.parse(savedFilters) as Partial<Filters>
        const merged = { ...defaultFilters, ...parsed }
        setFilters(merged)
        setDraftFilters(merged)
      }
    } catch {
      // ignore localStorage failures
    }
  }, [token])

  const trimmedQuery = useMemo(() => query.trim(), [query])

  async function runSearch(nextQuery: string) {
    const safeQuery = nextQuery.trim()
    if (!safeQuery) {
      setError('Type a keyword to search listings.')
      setResults([])
      setHasSearched(false)
      return
    }
    setLoading(true)
    setError('')
    setHasSearched(true)
    setLastQuery(safeQuery)
    try {
      const data = await searchListings(token, { query: safeQuery, city: userCity || undefined })
      setResults(Array.isArray(data.listings) ? data.listings : [])
      setPage(1)
      const updated = [safeQuery, ...recent.filter((q) => q !== safeQuery)].slice(0, MAX_RECENT)
      setRecent(updated)
      saveRecent(updated)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Search failed')
      setResults([])
    } finally {
      setLoading(false)
    }
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    void runSearch(query)
  }

  const filteredResults = useMemo(() => {
    const safeResults = Array.isArray(results) ? results : []
    const min = filters.minPrice ? Number(filters.minPrice) : null
    const max = filters.maxPrice ? Number(filters.maxPrice) : null
    return safeResults.filter((listing) => {
      if (filters.categoryId && listing.category_id !== filters.categoryId) return false
      if (filters.condition && listing.condition !== filters.condition) return false
      if (filters.city && listing.city.toLowerCase() !== filters.city.toLowerCase()) return false
      if (min !== null && Number(listing.price) < min) return false
      if (max !== null && Number(listing.price) > max) return false
      return true
    })
  }, [results, filters])

  const sortedResults = useMemo(() => {
    const list = [...filteredResults]
    if (filters.sort === 'price_low') {
      list.sort((a, b) => Number(a.price) - Number(b.price))
    } else if (filters.sort === 'price_high') {
      list.sort((a, b) => Number(b.price) - Number(a.price))
    } else {
      list.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    }
    return list
  }, [filteredResults, filters.sort])

  const totalPages = Math.max(1, Math.ceil(sortedResults.length / PAGE_SIZE))
  const pagedResults = sortedResults.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)
  const hasActiveFilters =
    filters.minPrice ||
    filters.maxPrice ||
    filters.categoryId ||
    filters.condition ||
    filters.city ||
    filters.sort !== 'newest'

  const categoryLabel =
    filters.categoryId && categories.find((c) => c.id === filters.categoryId)?.name
  const conditionLabel =
    filters.condition && LISTING_CONDITIONS.find((c) => c.value === filters.condition)?.label

  const mockPreviewListing: Listing = {
    id: 'preview_listing',
    seller_id: 'demo_seller',
    category_id: null,
    title: 'Vintage Desk Lamp',
    description:
      'A classic metal desk lamp with adjustable arm. Fully functional and in great condition.',
    condition: 'good',
    price: 1500,
    currency: 'INR',
    city: userCity || 'Your City',
    state: null,
    status: 'active',
    view_count: 0,
    created_at: new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
    images: [
      {
        id: 'preview_img_1',
        listing_id: 'preview_listing',
        image_url: 'https://placehold.co/800x600?text=Lamp+1',
        position: 0
      },
      {
        id: 'preview_img_2',
        listing_id: 'preview_listing',
        image_url: 'https://placehold.co/800x600?text=Lamp+2',
        position: 1
      }
    ]
  }

  useEffect(() => {
    setPage(1)
  }, [filters, results])

  return (
    <div className="search-page">
      <div className="search-header">
        <div>
          <h2 className="search-title">Search Listings</h2>
          <p className="search-subtitle">Find items for sale near you.</p>
        </div>
        <div className="search-header-actions">
          <button
            type="button"
            className="profile-edit-btn secondary"
            onClick={() => setSelected(mockPreviewListing)}
          >
            Preview Listing
          </button>
          <button type="button" className="profile-edit-btn secondary" onClick={() => {
            setDraftFilters(filters)
            setShowFilters(true)
          }}>
            Filters
          </button>
          <button type="button" className="back-link-btn" onClick={onBack}>
            <IconBack className="back-link-icon" aria-hidden />
            <span>Back to Profile</span>
          </button>
        </div>
      </div>

      <form className="search-bar" onSubmit={handleSubmit}>
        <div className="search-input-wrap">
          <IconSearch className="search-input-icon" aria-hidden />
          <input
            type="text"
            className="search-input"
            placeholder={`Search in ${userCity || 'your city'}`}
            value={query}
            onChange={(e) => {
              setQuery(e.target.value)
              if (error) setError('')
            }}
            onFocus={() => setShowRecent(true)}
            onBlur={() => {
              setTimeout(() => setShowRecent(false), 150)
            }}
          />
          <button type="submit" className="search-icon-btn" aria-label="Search">
            <IconSearch className="search-icon-btn-icon" aria-hidden />
          </button>
        </div>
        <button type="submit" className="profile-edit-btn primary">
          Search
        </button>
      </form>

      <div className="search-controls">
        <label className="search-sort">
          <span>Sort by</span>
          <select
            value={filters.sort}
            onChange={(e) => {
              const next = { ...filters, sort: e.target.value as SortOption }
              setFilters(next)
              setDraftFilters(next)
              localStorage.setItem(FILTERS_KEY, JSON.stringify(next))
              setFiltering(true)
              setTimeout(() => setFiltering(false), 150)
            }}
          >
            <option value="newest">Newest</option>
            <option value="price_low">Price: Low to High</option>
            <option value="price_high">Price: High to Low</option>
          </select>
        </label>
      </div>

      {hasActiveFilters && (
        <div className="search-applied">
          <div className="search-applied-title">Applied filters</div>
          <div className="search-applied-list">
            {filters.minPrice && (
              <button
                type="button"
                className="search-filter-chip"
                onClick={() => {
                  const next = { ...filters, minPrice: '' }
                  setFilters(next)
                  setDraftFilters(next)
                  localStorage.setItem(FILTERS_KEY, JSON.stringify(next))
                }}
              >
                Min {filters.minPrice}
                <span>x</span>
              </button>
            )}
            {filters.maxPrice && (
              <button
                type="button"
                className="search-filter-chip"
                onClick={() => {
                  const next = { ...filters, maxPrice: '' }
                  setFilters(next)
                  setDraftFilters(next)
                  localStorage.setItem(FILTERS_KEY, JSON.stringify(next))
                }}
              >
                Max {filters.maxPrice}
                <span>x</span>
              </button>
            )}
            {filters.categoryId && (
              <button
                type="button"
                className="search-filter-chip"
                onClick={() => {
                  const next = { ...filters, categoryId: '' }
                  setFilters(next)
                  setDraftFilters(next)
                  localStorage.setItem(FILTERS_KEY, JSON.stringify(next))
                }}
              >
                {categoryLabel || 'Category'}
                <span>x</span>
              </button>
            )}
            {filters.condition && (
              <button
                type="button"
                className="search-filter-chip"
                onClick={() => {
                  const next = { ...filters, condition: '' }
                  setFilters(next)
                  setDraftFilters(next)
                  localStorage.setItem(FILTERS_KEY, JSON.stringify(next))
                }}
              >
                {conditionLabel || 'Condition'}
                <span>x</span>
              </button>
            )}
            {filters.city && (
              <button
                type="button"
                className="search-filter-chip"
                onClick={() => {
                  const next = { ...filters, city: '' }
                  setFilters(next)
                  setDraftFilters(next)
                  localStorage.setItem(FILTERS_KEY, JSON.stringify(next))
                }}
              >
                City
                <span>x</span>
              </button>
            )}
            <button
              type="button"
              className="search-filter-clear"
              onClick={() => {
                setFilters(defaultFilters)
                setDraftFilters(defaultFilters)
                localStorage.setItem(FILTERS_KEY, JSON.stringify(defaultFilters))
              }}
            >
              Clear all
            </button>
          </div>
        </div>
      )}

      {showRecent && trimmedQuery.length === 0 && recent.length > 0 && (
        <div className="search-recent">
          <div className="search-recent-header">
            <div className="search-recent-title">Recent searches</div>
            <button
              type="button"
              className="search-recent-clear"
              onClick={() => {
                setRecent([])
                saveRecent([])
              }}
            >
              Clear all
            </button>
          </div>
          <div className="search-recent-list">
            {recent.map((item) => (
              <div key={item} className="search-chip">
                <button
                  type="button"
                  className="search-chip-label"
                  onClick={() => {
                    setQuery(item)
                    void runSearch(item)
                  }}
                >
                  {item}
                </button>
                <button
                  type="button"
                  className="search-chip-remove"
                  aria-label={`Remove ${item}`}
                  onClick={() => {
                    const updated = recent.filter((q) => q !== item)
                    setRecent(updated)
                    saveRecent(updated)
                  }}
                >
                  x
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {error && (
        <div className="search-error">
          <p className="error-message">{error}</p>
          {lastQuery && (
            <button
              type="button"
              className="profile-edit-btn secondary"
              onClick={() => void runSearch(lastQuery)}
            >
              Retry search
            </button>
          )}
        </div>
      )}

      {(loading || filtering) && (
        <div className="search-state">
          <div className="search-skeleton" />
          <div className="search-skeleton" />
          <div className="search-skeleton" />
        </div>
      )}

      {!loading && !filtering && hasSearched && results.length === 0 && !error && (
        <div className="search-empty">
          <h3>No matches yet</h3>
          <p>Try a different keyword or broaden your search.</p>
        </div>
      )}

      {!loading && !filtering && results.length > 0 && sortedResults.length === 0 && (
        <div className="search-empty">
          <h3>No listings match these filters</h3>
          <p>Try removing a filter or expanding the price range.</p>
        </div>
      )}

      {!loading && !filtering && sortedResults.length > 0 && (
        <div className="search-results">
          <div className="search-results-title">Results ({sortedResults.length})</div>
          <ul className="search-results-grid">
            {pagedResults.map((listing) => (
              <li
                key={listing.id}
                className="search-card"
                onClick={() => setSelected(listing)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') setSelected(listing)
                }}
              >
                <div className="search-card-image">
                  {listing.images?.[0] ? (
                    <img src={listing.images[0].image_url} alt={listing.title} />
                  ) : (
                    <div className="search-card-placeholder">No image</div>
                  )}
                </div>
                <div className="search-card-body">
                  <h3 className="search-card-title">{listing.title}</h3>
                  <p className="search-card-price">INR {Number(listing.price).toLocaleString()}</p>
                  <p className="search-card-meta">{listing.city}</p>
                </div>
              </li>
            ))}
          </ul>
          {totalPages > 1 && (
            <div className="search-pagination">
              <button
                type="button"
                className="search-page-btn"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                Previous
              </button>
              <span className="search-page-info">
                Page {page} of {totalPages}
              </span>
              <button
                type="button"
                className="search-page-btn"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              >
                Next
              </button>
            </div>
          )}
        </div>
      )}

      {selected && (
        <div className="search-modal-overlay" role="dialog" aria-modal="true">
          <div className="search-modal listing-details-modal">
            <ListingDetails
              listing={selected}
              onClose={() => setSelected(null)}
              onStartChat={
                onStartChat
                  ? async (listing) => {
                      await onStartChat(listing)
                      setSelected(null)
                    }
                  : undefined
              }
            />
          </div>
        </div>
      )}

      {showFilters && (
        <div className="search-modal-overlay" role="dialog" aria-modal="true">
          <div className="search-modal search-filter-modal">
            <div className="search-modal-header">
              <h3>Filters</h3>
              <button type="button" className="search-modal-close" onClick={() => setShowFilters(false)}>
                x
              </button>
            </div>
            <div className="search-filter-grid">
              <label>
                Min price
                <input
                  type="number"
                  value={draftFilters.minPrice}
                  onChange={(e) => setDraftFilters((prev) => ({ ...prev, minPrice: e.target.value }))}
                  placeholder="0"
                  min={0}
                />
              </label>
              <label>
                Max price
                <input
                  type="number"
                  value={draftFilters.maxPrice}
                  onChange={(e) => setDraftFilters((prev) => ({ ...prev, maxPrice: e.target.value }))}
                  placeholder="50000"
                  min={0}
                />
              </label>
              <label>
                Category
                <select
                  value={draftFilters.categoryId}
                  onChange={(e) => setDraftFilters((prev) => ({ ...prev, categoryId: e.target.value }))}
                >
                  <option value="">All</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Condition
                <select
                  value={draftFilters.condition}
                  onChange={(e) => setDraftFilters((prev) => ({ ...prev, condition: e.target.value }))}
                >
                  <option value="">All</option>
                  {LISTING_CONDITIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                City
                <input
                  type="text"
                  value={draftFilters.city}
                  onChange={(e) => setDraftFilters((prev) => ({ ...prev, city: e.target.value }))}
                  placeholder={userCity || 'City'}
                />
              </label>
            </div>
            <div className="search-filter-actions">
              <button
                type="button"
                className="profile-edit-btn secondary"
                onClick={() => {
                  setDraftFilters(defaultFilters)
                }}
              >
                Clear
              </button>
              <button
                type="button"
                className="profile-edit-btn primary"
                onClick={() => {
                  setFilters(draftFilters)
                  localStorage.setItem(FILTERS_KEY, JSON.stringify(draftFilters))
                  setFiltering(true)
                  setTimeout(() => setFiltering(false), 150)
                  setShowFilters(false)
                }}
              >
                Apply filters
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
