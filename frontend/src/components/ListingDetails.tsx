import { useEffect, useMemo, useState } from 'react'
import type { Listing } from '../types/listing'
import { LISTING_STATUS_LABELS } from '../types/listing'
import { IconBack, IconImage, IconLocation, IconUser } from './Icons'

interface ListingDetailsProps {
  listing: Listing
  onClose: () => void
  onBack?: () => void
  onStartChat?: (listing: Listing) => Promise<void>
  isFavorite?: boolean
  onToggleFavorite?: (listing: Listing, nextFavorite: boolean) => Promise<void>
  onReport?: (listing: Listing) => Promise<void>
}

function formatRelativeTime(iso: string): string {
  const created = new Date(iso).getTime()
  if (!created) return 'Just now'
  const diffMs = Date.now() - created
  const minutes = Math.floor(diffMs / 60000)
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes} min ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} hr ago`
  const days = Math.floor(hours / 24)
  return `${days} day${days === 1 ? '' : 's'} ago`
}

export default function ListingDetails({
  listing,
  onClose,
  onBack,
  onStartChat,
  isFavorite = false,
  onToggleFavorite,
  onReport
}: ListingDetailsProps) {
  const images = listing.images || []
  const [index, setIndex] = useState(0)
  const [favorite, setFavorite] = useState(isFavorite)
  const [touchStart, setTouchStart] = useState<number | null>(null)
  const [chatting, setChatting] = useState(false)
  const [chatError, setChatError] = useState('')
  const [favoriteLoading, setFavoriteLoading] = useState(false)
  const [favoriteError, setFavoriteError] = useState('')
  const [reporting, setReporting] = useState(false)
  const [reportError, setReportError] = useState('')
  const [reportMessage, setReportMessage] = useState('')

  const currentImage = useMemo(() => images[index], [images, index])
  const isUnavailable = listing.status === 'deleted'
  const statusLabel = LISTING_STATUS_LABELS[listing.status] || listing.status

  useEffect(() => {
    setFavorite(isFavorite)
  }, [isFavorite])

  const goNext = () => {
    if (images.length === 0) return
    setIndex((prev) => (prev + 1) % images.length)
  }

  const goPrev = () => {
    if (images.length === 0) return
    setIndex((prev) => (prev - 1 + images.length) % images.length)
  }

  const handleStartChat = async () => {
    if (!onStartChat) return
    setChatting(true)
    setChatError('')
    try {
      await onStartChat(listing)
    } catch (error: unknown) {
      setChatError(error instanceof Error ? error.message : 'Failed to start chat.')
    } finally {
      setChatting(false)
    }
  }

  const handleToggleFavorite = async () => {
    if (!onToggleFavorite || favoriteLoading) return
    setFavoriteLoading(true)
    setFavoriteError('')
    const nextFavorite = !favorite
    try {
      await onToggleFavorite(listing, nextFavorite)
      setFavorite(nextFavorite)
    } catch (error: unknown) {
      setFavoriteError(error instanceof Error ? error.message : 'Failed to update favorite.')
    } finally {
      setFavoriteLoading(false)
    }
  }

  const handleReport = async () => {
    if (!onReport || reporting) return
    setReporting(true)
    setReportError('')
    setReportMessage('')
    try {
      await onReport(listing)
      setReportMessage('Listing reported.')
    } catch (error: unknown) {
      setReportError(error instanceof Error ? error.message : 'Failed to report listing.')
    } finally {
      setReporting(false)
    }
  }

  return (
    <div className="listing-details">
      <div className="listing-details-header">
        <button type="button" className="back-link-btn" onClick={onBack || onClose}>
          <IconBack className="back-link-icon" aria-hidden />
          <span>Back</span>
        </button>
        <button type="button" className="listing-details-close" onClick={onClose}>
          x
        </button>
      </div>

      {isUnavailable && (
        <div className="listing-details-unavailable">
          This listing is no longer available.
        </div>
      )}

      <div className="listing-details-gallery">
        {currentImage ? (
          <div
            className="listing-details-image-wrap"
            onTouchStart={(e) => setTouchStart(e.touches[0]?.clientX ?? null)}
            onTouchEnd={(e) => {
              if (touchStart === null) return
              const end = e.changedTouches[0]?.clientX ?? touchStart
              const delta = end - touchStart
              if (Math.abs(delta) > 40) {
                if (delta < 0) goNext()
                if (delta > 0) goPrev()
              }
              setTouchStart(null)
            }}
          >
            <img src={currentImage.image_url} alt={listing.title} />
            {images.length > 1 && (
              <>
                <button type="button" className="listing-details-nav prev" onClick={goPrev}>
                  {'<'}
                </button>
                <button type="button" className="listing-details-nav next" onClick={goNext}>
                  {'>'}
                </button>
              </>
            )}
          </div>
        ) : (
          <div className="listing-details-placeholder">
            <IconImage className="listing-details-placeholder-icon" aria-hidden />
            <p>No images</p>
          </div>
        )}

        {images.length > 1 && (
          <div className="listing-details-thumbs">
            {images.map((img, i) => (
              <button
                key={img.id}
                type="button"
                className={`listing-details-thumb ${i === index ? 'active' : ''}`}
                onClick={() => setIndex(i)}
              >
                <img src={img.image_url} alt={`Photo ${i + 1}`} />
              </button>
            ))}
          </div>
        )}
      </div>

      <div className="listing-details-body">
        <div className="listing-details-title-row">
          <h2>{listing.title}</h2>
          <span className={`listing-details-status ${listing.status}`}>{statusLabel}</span>
        </div>
        <p className="listing-details-price">INR {Number(listing.price).toLocaleString()}</p>
        <p className="listing-details-meta">
          <IconLocation className="listing-details-meta-icon" aria-hidden />
          {listing.city}
          {listing.state ? `, ${listing.state}` : ''}
          <span className="listing-details-dot">·</span>
          {formatRelativeTime(listing.created_at)}
        </p>
        {listing.condition && (
          <p className="listing-details-condition">
            Condition: <strong>{listing.condition}</strong>
          </p>
        )}
        <p className="listing-details-desc">{listing.description}</p>

        <div className="listing-details-seller">
          <IconUser className="listing-details-seller-icon" aria-hidden />
          <div>
            <div className="listing-details-seller-name">Seller</div>
            <div className="listing-details-seller-meta">{listing.seller_id}</div>
          </div>
        </div>

        <div className="listing-details-actions">
          <button
            type="button"
            className="profile-edit-btn primary"
            onClick={handleStartChat}
            disabled={chatting || !onStartChat}
          >
            {chatting ? 'Starting chat...' : 'Chat with Seller'}
          </button>
          <button
            type="button"
            className="profile-edit-btn secondary"
            onClick={() => {
              void handleToggleFavorite()
            }}
            disabled={favoriteLoading || !onToggleFavorite}
          >
            {favoriteLoading ? 'Saving...' : favorite ? 'Unfavorite' : 'Favorite'}
          </button>
          <button
            type="button"
            className="profile-edit-btn secondary"
            onClick={() => {
              void handleReport()
            }}
            disabled={reporting || !onReport}
          >
            {reporting ? 'Reporting...' : 'Report'}
          </button>
        </div>
        {favoriteError && (
          <div className="listing-details-chat-error">
            <span>{favoriteError}</span>
          </div>
        )}
        {reportError && (
          <div className="listing-details-chat-error">
            <span>{reportError}</span>
          </div>
        )}
        {reportMessage && (
          <div className="message">
            <span>{reportMessage}</span>
          </div>
        )}
        {chatError && (
          <div className="listing-details-chat-error">
            <span>{chatError}</span>
            <button type="button" className="profile-edit-btn secondary" onClick={handleStartChat}>
              Retry
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
