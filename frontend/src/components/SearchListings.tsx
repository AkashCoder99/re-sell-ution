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
const MAX_RECENT = 6

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

  useEffect(() => {
    setRecent(loadRecent())
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
    try {
      const data = await searchListings(token, { query: safeQuery, city: userCity || undefined })
      setResults(data.listings)
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
          />
        </div>
        <button type="submit" className="profile-edit-btn primary">
          Search
        </button>
      </form>

      {recent.length > 0 && (
        <div className="search-recent">
          <div className="search-recent-title">Recent searches</div>
          <div className="search-recent-list">
            {recent.map((item) => (
              <button
                key={item}
                type="button"
                className="search-chip"
                onClick={() => {
                  setQuery(item)
                  void runSearch(item)
                }}
              >
                {item}
              </button>
            ))}
          </div>
        </div>
      )}

      {error && <p className="error-message">{error}</p>}

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
            {results.map((listing) => (
              <li key={listing.id} className="search-card">
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
        </div>
      )}
    </div>
  )
}
