/**
 * BrowseListings — public feed filtered by city and category (completes F5)
 */
import { useEffect, useState } from 'react'
import type { Listing, Category } from '../types/listing'
import { getCategories } from '../api/listings'
import { POPULAR_CITIES } from '../utils/constants'

interface BrowseListingsProps {
  token: string
  userCity?: string
  onBack: () => void
}

interface BrowseResponse {
  listings: Listing[]
  total: number
  page: number
  total_pages: number
}

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true'

async function fetchPublicListings(params: {
  city?: string
  category_id?: string
  page: number
  limit: number
  token: string
}): Promise<BrowseResponse> {
  if (USE_MOCK) {
    // Return mock data for demo mode
    await new Promise((r) => setTimeout(r, 300))
    const sampleListings = [
      { title: 'Apple MacBook Pro 14" M2', price: 1299, city: params.city || 'Gainesville', condition: 'like_new', image: 'https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=400&h=300&fit=crop' },
      { title: 'Sony WH-1000XM5 Headphones', price: 220, city: params.city || 'Gainesville', condition: 'good', image: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&h=300&fit=crop' },
      { title: 'IKEA MALM Queen Bed Frame', price: 85, city: params.city || 'Gainesville', condition: 'good', image: 'https://images.unsplash.com/photo-1555041469-a586c61ea9bc?w=400&h=300&fit=crop' },
      { title: 'Nike Air Force 1 Low — Size 10', price: 95, city: params.city || 'Gainesville', condition: 'new', image: 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400&h=300&fit=crop' },
      { title: 'UF CS Textbook Bundle', price: 45, city: params.city || 'Gainesville', condition: 'good', image: 'https://images.unsplash.com/photo-1544716278-ca5e3f4abd8c?w=400&h=300&fit=crop' },
      { title: 'iPad Air 5th Gen + Apple Pencil', price: 550, city: params.city || 'Gainesville', condition: 'like_new', image: 'https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?w=400&h=300&fit=crop' },
      { title: 'Electric Standing Desk 55"', price: 180, city: params.city || 'Gainesville', condition: 'good', image: 'https://images.unsplash.com/photo-1593642632559-0c6d3fc62b89?w=400&h=300&fit=crop' },
      { title: 'Trek FX3 Hybrid Bike — Medium', price: 420, city: params.city || 'Gainesville', condition: 'like_new', image: 'https://images.unsplash.com/photo-1485965120184-e220f721d03e?w=400&h=300&fit=crop' },
    ]
    const mockItems: Listing[] = sampleListings.map((s, i) => ({
      id: `browse_${i}`,
      seller_id: 'mock_seller',
      category_id: null,
      title: s.title,
      description: 'Available for pickup in ' + s.city + '. Message for details.',
      condition: s.condition as Listing['condition'],
      price: s.price,
      currency: 'USD',
      city: s.city,
      state: 'Florida',
      status: 'active' as const,
      view_count: Math.floor(Math.random() * 60) + 5,
      created_at: new Date(Date.now() - i * 86400000).toISOString(),
      updated_at: new Date(Date.now() - i * 86400000).toISOString(),
      images: [{ id: `img_browse_${i}`, listing_id: `browse_${i}`, image_url: s.image, position: 0 }]
    }))
    return { listings: mockItems, total: mockItems.length, page: 1, total_pages: 1 }
  }

  const sp = new URLSearchParams()
  if (params.city) sp.set('city', params.city)
  if (params.category_id) sp.set('category_id', params.category_id)
  sp.set('page', String(params.page))
  sp.set('limit', String(params.limit))

  const res = await fetch(`${API_BASE}/api/v1/listings?${sp.toString()}`, {
    headers: { Authorization: `Bearer ${params.token}` }
  })
  if (!res.ok) throw new Error('Failed to load listings')
  return res.json() as Promise<BrowseResponse>
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
      .then((r) => setCategories(r.categories))
      .catch(() => {})
  }, [token])

  useEffect(() => {
    setLoading(true)
    setError('')
    fetchPublicListings({ city: selectedCity, category_id: selectedCategory, page, limit: 12, token })
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
      </div>

      {error && <p className="error-message">{error}</p>}

      {loading ? (
        <p className="browse-loading">Loading listings...</p>
      ) : listings.length === 0 ? (
        <p className="browse-empty">No listings found for the selected filters.</p>
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
                <h3 className="browse-card-title">{listing.title}</h3>
                <p className="browse-card-price">
                  {listing.currency} {listing.price.toLocaleString()}
                </p>
                <p className="browse-card-meta">
                  {listing.city} · {listing.condition.replace('_', ' ')}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="browse-pagination">
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
            className="browse-page-btn"
          >
            ← Prev
          </button>
          <span>Page {page} of {totalPages}</span>
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
            className="browse-page-btn"
          >
            Next →
          </button>
        </div>
      )}
    </div>
  )
}
