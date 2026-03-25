import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import type { Listing } from '../types/listing'
import { searchListings } from '../api/listings'
import { IconBack, IconSearch } from './Icons'

interface SearchListingsProps {
  token: string
  userCity: string
  onBack: () => void
}

const RECENT_KEY = 'resellution_recent_searches'
const PREFILL_KEY = 'resellution_search_prefill'
const MAX_RECENT = 6
const PAGE_SIZE = 8

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

export default function SearchListings({ token, userCity, onBack }: SearchListingsProps) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<Listing[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [hasSearched, setHasSearched] = useState(false)
  const [recent, setRecent] = useState<string[]>([])
  const [page, setPage] = useState(1)
  const [lastQuery, setLastQuery] = useState('')
  const [showRecent, setShowRecent] = useState(false)
  const [selected, setSelected] = useState<Listing | null>(null)

  useEffect(() => {
    setRecent(loadRecent())
    try {
      const prefill = localStorage.getItem(PREFILL_KEY) || ''
      if (prefill) {
        setQuery(prefill)
        localStorage.removeItem(PREFILL_KEY)
        void runSearch(prefill)
      }
    } catch {
      // ignore localStorage failures
    }
  }, [])

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
      setResults(data.listings)
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

  const totalPages = Math.max(1, Math.ceil(results.length / PAGE_SIZE))
  const pagedResults = results.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  return (
    <div className="search-page">
      <div className="search-header">
        <div>
          <h2 className="search-title">Search Listings</h2>
          <p className="search-subtitle">Find items for sale near you.</p>
        </div>
        <button type="button" className="back-link-btn" onClick={onBack}>
          <IconBack className="back-link-icon" aria-hidden />
          <span>Back to Profile</span>
        </button>
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
                  ×
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

      {loading && (
        <div className="search-state">
          <div className="search-skeleton" />
          <div className="search-skeleton" />
          <div className="search-skeleton" />
        </div>
      )}

      {!loading && hasSearched && results.length === 0 && !error && (
        <div className="search-empty">
          <h3>No matches yet</h3>
          <p>Try a different keyword or broaden your search.</p>
        </div>
      )}

      {!loading && results.length > 0 && (
        <div className="search-results">
          <div className="search-results-title">Results ({results.length})</div>
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
          <div className="search-modal">
            <div className="search-modal-header">
              <h3>{selected.title}</h3>
              <button type="button" className="search-modal-close" onClick={() => setSelected(null)}>
                ×
              </button>
            </div>
            <p className="search-modal-price">INR {Number(selected.price).toLocaleString()}</p>
            <p className="search-modal-meta">{selected.city}</p>
            <p className="search-modal-desc">{selected.description}</p>
            <div className="search-modal-actions">
              <button type="button" className="profile-edit-btn secondary" onClick={() => setSelected(null)}>
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
