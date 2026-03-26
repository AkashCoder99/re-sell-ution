/**
 * BrowseListings — public feed filtered by city and category (completes F5)
 */
import { useEffect, useState } from 'react'
import type { Listing, Category } from '../types/listing'
import { getCategories, getPublicListings } from '../api/listings'
import { POPULAR_CITIES } from '../utils/constants'
import { LISTING_CONDITIONS } from '../types/listing'

interface BrowseListingsProps {
  token: string
  userCity?: string
  onBack: () => void
}

export default function BrowseListings({ token, userCity, onBack }: BrowseListingsProps) {
  const [listings, setListings] = useState<Listing[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [selectedCity, setSelectedCity] = useState(userCity || '')
  const [selectedCategory, setSelectedCategory] = useState('')
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    getCategories(token)
      .then((r) => setCategories(Array.isArray(r.categories) ? r.categories : []))
      .catch(() => {})
  }, [token])

  useEffect(() => {
    setLoading(true)
    setError('')
    getPublicListings(token, { city: selectedCity, category_id: selectedCategory, page, limit: 12 })
      .then((r) => {
        setListings(r.listings)
        setTotalPages(r.total_pages)
      })
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'Failed to load listings'))
      .finally(() => setLoading(false))
  }, [selectedCity, selectedCategory, page, token])

  function handleCityChange(e: React.ChangeEvent<HTMLSelectElement>) {
    setSelectedCity(e.target.value)
    setPage(1)
  }

  function handleCategoryChange(e: React.ChangeEvent<HTMLSelectElement>) {
    setSelectedCategory(e.target.value)
    setPage(1)
  }

  function getCategoryName(id: string | null) {
    if (!id) return null
    return categories.find((c) => c.id === id)?.name ?? null
  }

  function getConditionLabel(condition: string) {
    return LISTING_CONDITIONS.find((c) => c.value === condition)?.label ?? condition
  }

  return (
    <div className="browse-listings">
      <div className="browse-header">
        <button type="button" className="back-link-btn" onClick={onBack}>
          ← Back
        </button>
        <h2>Browse Listings</h2>
      </div>

      <div className="browse-filters">
        <select
          value={selectedCity}
          onChange={handleCityChange}
          className="browse-filter-select"
          aria-label="Filter by city"
        >
          <option value="">All Cities</option>
          {POPULAR_CITIES.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>

        <select
          value={selectedCategory}
          onChange={handleCategoryChange}
          className="browse-filter-select"
          aria-label="Filter by category"
        >
          <option value="">All Categories</option>
          {categories.map((cat) => (
            <option key={cat.id} value={cat.id}>{cat.name}</option>
          ))}
        </select>

        {(selectedCity || selectedCategory) && (
          <button
            type="button"
            className="browse-filter-clear"
            onClick={() => { setSelectedCity(''); setSelectedCategory(''); setPage(1) }}
          >
            ✕ Clear filters
          </button>
        )}
      </div>

      {(selectedCity || selectedCategory) && (
        <p className="browse-active-filters">
          Showing{selectedCategory ? ` ${getCategoryName(selectedCategory) ?? ''}` : ''} listings
          {selectedCity ? ` in ${selectedCity}` : ''}
          {' '}({listings.length} result{listings.length !== 1 ? 's' : ''})
        </p>
      )}

      {error && <p className="error-message">{error}</p>}

      {loading ? (
        <p className="browse-loading">Loading listings...</p>
      ) : listings.length === 0 ? (
        <div className="browse-empty">
          <p>No listings found{selectedCategory ? ` in ${getCategoryName(selectedCategory)}` : ''}{selectedCity ? ` in ${selectedCity}` : ''}.</p>
          {(selectedCity || selectedCategory) && (
            <button
              type="button"
              className="browse-filter-clear"
              onClick={() => { setSelectedCity(''); setSelectedCategory(''); setPage(1) }}
            >
              Clear filters
            </button>
          )}
        </div>
      ) : (
        <div className="browse-grid">
          {listings.map((listing) => (
            <div key={listing.id} className="browse-card">
              {listing.images && listing.images.length > 0 ? (
                <img
                  src={listing.images[0].image_url}
                  alt={listing.title}
                  className="browse-card-img"
                />
              ) : (
                <div className="browse-card-img-placeholder">No Image</div>
              )}
              <div className="browse-card-body">
                {getCategoryName(listing.category_id) && (
                  <span className="browse-card-category">{getCategoryName(listing.category_id)}</span>
                )}
                <h3 className="browse-card-title">{listing.title}</h3>
                <p className="browse-card-price">${listing.price.toLocaleString()}</p>
                <p className="browse-card-meta">
                  📍 {listing.city} · {getConditionLabel(listing.condition)}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="browse-pagination">
          <button type="button" disabled={page <= 1} onClick={() => setPage((p) => p - 1)} className="browse-page-btn">← Prev</button>
          <span>Page {page} of {totalPages}</span>
          <button type="button" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)} className="browse-page-btn">Next →</button>
        </div>
      )}
    </div>
  )
}
