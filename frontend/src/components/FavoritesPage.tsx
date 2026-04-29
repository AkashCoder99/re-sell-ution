import { useEffect, useState } from 'react'
import type { Listing } from '../types/listing'
import { getFavoriteListings, removeFavorite } from '../api/listings'
import { IconBack } from './Icons'

interface FavoritesPageProps {
  token: string
  onBack: () => void
}

export default function FavoritesPage({ token, onBack }: FavoritesPageProps) {
  const [favorites, setFavorites] = useState<Listing[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [removingId, setRemovingId] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError('')

    getFavoriteListings(token)
      .then((res) => {
        if (cancelled) return
        setFavorites(res.listings)
      })
      .catch((err: unknown) => {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'Failed to load favorites.')
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [token])

  async function handleRemove(listingId: string) {
    setRemovingId(listingId)
    setError('')
    try {
      await removeFavorite(token, listingId)
      setFavorites((prev) => prev.filter((item) => item.id !== listingId))
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to remove favorite.')
    } finally {
      setRemovingId(null)
    }
  }

  return (
    <div className="favorites-page">
      <div className="browse-header">
        <button type="button" className="back-link-btn" onClick={onBack}>
          <IconBack className="back-link-icon" aria-hidden />
          <span>Back to Profile</span>
        </button>
        <h2>Saved Listings</h2>
      </div>

      {error && <p className="error-message">{error}</p>}

      {loading ? (
        <p className="browse-loading">Loading favorites...</p>
      ) : favorites.length === 0 ? (
        <div className="browse-empty">
          <p>No saved listings yet.</p>
        </div>
      ) : (
        <div className="browse-grid">
          {favorites.map((listing) => (
            <div key={listing.id} className="browse-card">
              {listing.images && listing.images.length > 0 ? (
                <img src={listing.images[0].image_url} alt={listing.title} className="browse-card-img" />
              ) : (
                <div className="browse-card-img-placeholder">No Image</div>
              )}
              <div className="browse-card-body">
                <h3 className="browse-card-title">{listing.title}</h3>
                <p className="browse-card-price">INR {Number(listing.price).toLocaleString()}</p>
                <p className="browse-card-meta">{listing.city}</p>
                <button
                  type="button"
                  className="profile-edit-btn secondary"
                  onClick={() => {
                    void handleRemove(listing.id)
                  }}
                  disabled={removingId === listing.id}
                >
                  {removingId === listing.id ? 'Removing...' : 'Remove'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
