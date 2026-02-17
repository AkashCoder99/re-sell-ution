/**
 * F14 — My listings dashboard: Active / Sold / Drafts tabs, cards, quick actions, empty states, pagination
 * F20 — Mark as sold + optional buyer selection
 */

import { useState, useEffect, useCallback } from 'react'
import type { Listing, ListingStatus } from '../types/listing'
import { LISTING_STATUS_LABELS } from '../types/listing'
import { getMyListings, updateListingStatus, deleteListing } from '../api/listings'
import { IconListings } from './Icons'

type TabStatus = 'active' | 'sold' | 'draft'

const PAGE_SIZE = 10

interface MyListingsDashboardProps {
  token: string
  onEdit?: (listing: Listing) => void
  onMarkSold?: (listing: Listing) => void
  refreshTrigger?: number
}

export default function MyListingsDashboard({
  token,
  onEdit,
  onMarkSold,
  refreshTrigger = 0
}: MyListingsDashboardProps) {
  const [tab, setTab] = useState<TabStatus>('active')
  const [listings, setListings] = useState<Listing[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [markSoldListing, setMarkSoldListing] = useState<Listing | null>(null)
  const [soldToUserId, setSoldToUserId] = useState('')

  const fetchListings = useCallback(
    async (status: TabStatus, pageNum: number) => {
      setLoading(true)
      setError('')
      try {
        const res = await getMyListings(token, {
          status: status === 'draft' ? 'draft' : status,
          page: pageNum,
          limit: PAGE_SIZE
        })
        setListings(res.listings)
        setTotal(res.total)
        setPage(res.page)
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : 'Failed to load listings')
        setListings([])
        setTotal(0)
      } finally {
        setLoading(false)
      }
    },
    [token]
  )

  useEffect(() => {
    fetchListings(tab, page)
  }, [tab, page, refreshTrigger, fetchListings])

  const handleDelete = async (listing: Listing) => {
    if (!window.confirm('Delete this listing? This cannot be undone.')) return
    setActionLoading(listing.id)
    try {
      await deleteListing(token, listing.id)
      setListings((prev) => prev.filter((l) => l.id !== listing.id))
      setTotal((t) => Math.max(0, t - 1))
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Delete failed')
    } finally {
      setActionLoading(null)
    }
  }

  const handleMarkSoldClick = (listing: Listing) => {
    setMarkSoldListing(listing)
    setSoldToUserId('')
  }

  const handleMarkSoldConfirm = async () => {
    if (!markSoldListing) return
    setActionLoading(markSoldListing.id)
    try {
      await updateListingStatus(token, markSoldListing.id, {
        status: 'sold',
        ...(soldToUserId ? { sold_to_user_id: soldToUserId } : {})
      })
      setMarkSoldListing(null)
      setSoldToUserId('')
      setListings((prev) => prev.filter((l) => l.id !== markSoldListing.id))
      setTotal((t) => Math.max(0, t - 1))
      onMarkSold?.(markSoldListing)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to mark as sold')
    } finally {
      setActionLoading(null)
    }
  }

  const totalPages = Math.ceil(total / PAGE_SIZE) || 1

  return (
    <div className="my-listings-dashboard">
      <h2 className="my-listings-heading">
        <IconListings className="my-listings-heading-icon" aria-hidden />
        My Listings
      </h2>

      <div className="my-listings-tabs">
        {(['active', 'sold', 'draft'] as TabStatus[]).map((t) => (
          <button
            key={t}
            type="button"
            className={`my-listings-tab ${tab === t ? 'active' : ''}`}
            onClick={() => setTab(t)}
          >
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {error && <p className="error-message">{error}</p>}

      {loading ? (
        <p className="my-listings-loading">Loading...</p>
      ) : listings.length === 0 ? (
        <div className="my-listings-empty">
          <p className="my-listings-empty-title">
            {tab === 'active' && 'No active listings'}
            {tab === 'sold' && 'No sold items yet'}
            {tab === 'draft' && 'No drafts'}
          </p>
          <p className="my-listings-empty-hint">
            {tab === 'active' && 'Create a listing to start selling.'}
            {tab === 'sold' && 'Items you mark as sold will appear here.'}
            {tab === 'draft' && 'Save a listing as draft to finish later.'}
          </p>
        </div>
      ) : (
        <>
          <ul className="my-listings-grid">
            {listings.map((listing) => (
              <li key={listing.id} className="my-listings-card">
                <div className="my-listings-card-image">
                  {listing.images?.[0] ? (
                    <img
                      src={listing.images[0].image_url}
                      alt={listing.title}
                    />
                  ) : (
                    <div className="my-listings-card-placeholder">No image</div>
                  )}
                </div>
                <div className="my-listings-card-body">
                  <h3 className="my-listings-card-title">{listing.title}</h3>
                  <p className="my-listings-card-price">₹{Number(listing.price).toLocaleString()}</p>
                  <p className="my-listings-card-meta">
                    {LISTING_STATUS_LABELS[listing.status as ListingStatus]} · {listing.city}
                  </p>
                  <div className="my-listings-card-actions">
                    {listing.status === 'active' && (
                      <>
                        {onMarkSold && (
                          <button
                            type="button"
                            className="my-listings-btn sold"
                            onClick={() => handleMarkSoldClick(listing)}
                            disabled={!!actionLoading}
                          >
                            Mark sold
                          </button>
                        )}
                        {onEdit && (
                          <button
                            type="button"
                            className="my-listings-btn"
                            onClick={() => onEdit(listing)}
                            disabled={!!actionLoading}
                          >
                            Edit
                          </button>
                        )}
                        <button
                          type="button"
                          className="my-listings-btn danger"
                          onClick={() => handleDelete(listing)}
                          disabled={actionLoading === listing.id}
                        >
                          {actionLoading === listing.id ? 'Deleting...' : 'Delete'}
                        </button>
                      </>
                    )}
                    {listing.status === 'sold' && (
                      <span className="my-listings-badge sold">Sold</span>
                    )}
                  </div>
                </div>
              </li>
            ))}
          </ul>

          {totalPages > 1 && (
            <div className="my-listings-pagination">
              <button
                type="button"
                className="my-listings-page-btn"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                Previous
              </button>
              <span className="my-listings-page-info">
                Page {page} of {totalPages} ({total} total)
              </span>
              <button
                type="button"
                className="my-listings-page-btn"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
              </button>
            </div>
          )}
        </>
      )}

      {markSoldListing && (
        <div className="my-listings-modal-overlay" role="dialog" aria-modal="true">
          <div className="my-listings-modal">
            <h3>Mark as sold</h3>
            <p className="my-listings-modal-title">{markSoldListing.title}</p>
            <label>
              Buyer (optional) — select from chats or leave blank
              <input
                type="text"
                placeholder="Buyer user ID (optional)"
                value={soldToUserId}
                onChange={(e) => setSoldToUserId(e.target.value)}
              />
            </label>
            <div className="button-group">
              <button
                type="button"
                className="secondary"
                onClick={() => {
                  setMarkSoldListing(null)
                  setSoldToUserId('')
                }}
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleMarkSoldConfirm}
                disabled={actionLoading === markSoldListing.id}
              >
                {actionLoading === markSoldListing.id ? 'Updating...' : 'Mark as sold'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
