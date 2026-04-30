/**
 * F12 - Create listing: multi-step flow (basic info -> details -> photos -> review)
 */

import { useState, useEffect, useMemo } from 'react'
import type { ChangeEvent, FormEvent } from 'react'
import type { CreateListingDraft, ListingCondition, Listing } from '../types/listing'
import { LISTING_CONDITIONS } from '../types/listing'
import {
  validateListingTitle,
  validateListingDescription,
  validateListingPrice
} from '../utils/validation'
import PhotoUpload, { type PhotoItem } from './PhotoUpload'
import type { Category } from '../types/listing'
import {
  createListing,
  deleteListing,
  getCategories,
  updateListing,
  updateListingStatus,
  uploadListingImageFile
} from '../api/listings'
import { IconAddListing } from './Icons'

const STEPS = ['basic', 'details', 'photos', 'review'] as const
type Step = (typeof STEPS)[number]

const defaultDraft: CreateListingDraft = {
  title: '',
  description: '',
  condition: 'good',
  price: 0,
  currency: 'INR',
  city: '',
  state: '',
  category_id: null,
  image_urls: []
}

interface CreateListingProps {
  token: string
  userCity: string
  onSuccess: () => void
  onCancel: () => void
  initialListing?: Listing | null
}

export default function CreateListing({
  token,
  userCity,
  onSuccess,
  onCancel,
  initialListing = null
}: CreateListingProps) {
  const isEditMode = Boolean(initialListing)
  const [step, setStep] = useState<Step>('basic')
  const [draft, setDraft] = useState<CreateListingDraft>(() => ({
    ...defaultDraft,
    title: initialListing?.title || '',
    description: initialListing?.description || '',
    condition: initialListing?.condition || defaultDraft.condition,
    price: initialListing?.price || 0,
    currency: initialListing?.currency || defaultDraft.currency,
    city: initialListing?.city || userCity,
    state: initialListing?.state || '',
    category_id: initialListing?.category_id || null
  }))
  const [photos, setPhotos] = useState<PhotoItem[]>(
    () =>
      initialListing?.images?.map((img) => ({
        id: img.id,
        preview: img.image_url,
        url: img.image_url,
        progress: 100,
        status: 'done' as const
      })) || []
  )
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(false)
  const [stepLoading, setStepLoading] = useState(false)
  const [createdListingId, setCreatedListingId] = useState<string | null>(initialListing?.id || null)
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    getCategories(token)
      .then((res) => setCategories(Array.isArray(res.categories) ? res.categories : []))
      .catch(() => setCategories([]))
  }, [token])

  const safeCategories = Array.isArray(categories) ? categories : []

  const stepIndex = STEPS.indexOf(step)
  const progressPercent = ((stepIndex + 1) / STEPS.length) * 100
  const initialSignature = useMemo(() => {
    const base = initialListing
      ? {
          title: initialListing.title,
          description: initialListing.description,
          condition: initialListing.condition,
          price: Number(initialListing.price),
          currency: initialListing.currency,
          city: initialListing.city,
          state: initialListing.state || '',
          category_id: initialListing.category_id || null
        }
      : {
          ...defaultDraft,
          city: userCity,
          price: 0
        }
    const basePhotos = initialListing?.images?.map((img) => img.image_url) || []
    return JSON.stringify({
      ...base,
      photos: basePhotos
    })
  }, [initialListing, userCity])

  const currentSignature = useMemo(
    () =>
      JSON.stringify({
        title: draft.title,
        description: draft.description,
        condition: draft.condition,
        price: Number(draft.price),
        currency: draft.currency,
        city: draft.city,
        state: draft.state,
        category_id: draft.category_id || null,
        photos: photos.map((p) => p.url || p.preview || '')
      }),
    [draft, photos]
  )

  const hasUnsavedChanges = useMemo(() => {
    const formChanged = currentSignature !== initialSignature
    return formChanged || (!isEditMode && createdListingId !== null) || step !== 'basic'
  }, [currentSignature, initialSignature, isEditMode, createdListingId, step])

  useEffect(() => {
    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      if (!hasUnsavedChanges) return
      event.preventDefault()
      event.returnValue = ''
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [hasUnsavedChanges])

  const validateBasic = (): boolean => {
    const errs: Record<string, string> = {}
    const titleError = validateListingTitle(draft.title)
    if (titleError) errs.title = titleError
    const cityError = draft.city.trim() ? null : 'City is required'
    if (cityError) errs.city = cityError
    setFieldErrors(errs)
    return Object.keys(errs).length === 0
  }

  const validateDetails = (): boolean => {
    const errs: Record<string, string> = {}
    const descriptionError = validateListingDescription(draft.description)
    if (descriptionError) errs.description = descriptionError
    const priceError = validateListingPrice(draft.price)
    if (priceError) errs.price = priceError
    setFieldErrors(errs)
    return Object.keys(errs).length === 0
  }

  const goNext = async () => {
    setError('')
    setFieldErrors({})
    if (step === 'basic' && !validateBasic()) return
    if (step === 'details' && !validateDetails()) return

    // details -> photos: create listing (no images yet)
    if (step === 'details') {
      if (createdListingId) {
        setStep('photos')
        return
      }

      setStepLoading(true)
      try {
        const res = await createListing(token, {
          title: draft.title.trim(),
          description: draft.description.trim(),
          condition: draft.condition as ListingCondition,
          price: Number(draft.price),
          currency: draft.currency,
          city: draft.city.trim(),
          state: draft.state.trim() || undefined,
          category_id: draft.category_id || undefined,
          status: 'draft',
          image_urls: []
        })
        setCreatedListingId(res.listing.id)
        setStep('photos')
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : 'Failed to create listing')
      } finally {
        setStepLoading(false)
      }
      return
    }

    // photos -> review: upload images (persist listing_images) then go to review
    if (step === 'photos') {
      if (!createdListingId) {
        setError('Listing was not created yet. Please go back and retry.')
        return
      }

      const pendingWithIndex = photos
        .map((p, idx) => ({ p, idx }))
        .filter(({ p }) => p.status === 'pending' && p.file)

      if (pendingWithIndex.length === 0) {
        setStep('review')
        return
      }

      setStepLoading(true)
      try {
        for (const { p } of pendingWithIndex) {
          if (!p.file) continue

          setPhotos((prev) =>
            prev.map((x) =>
              x.id === p.id ? { ...x, status: 'uploading' as const, progress: 50, error: undefined } : x
            )
          )

          // Let the backend pick the next available position to avoid collisions.
          const res = await uploadListingImageFile(token, createdListingId, p.file)

          setPhotos((prev) =>
            prev.map((x) =>
              x.id === p.id
                ? { ...x, status: 'done' as const, progress: 100, url: res.image.image_url, error: undefined }
                : x
            )
          )
        }

        setStep('review')
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : 'Photo upload failed'
        setError(message)
        setPhotos((prev) =>
          prev.map((x) =>
            x.status === 'pending' ? { ...x, status: 'failed' as const, error: message } : x
          )
        )
      } finally {
        setStepLoading(false)
      }
      return
    }

    const nextIdx = stepIndex + 1
    if (nextIdx < STEPS.length) setStep(STEPS[nextIdx])
  }

  const goPrev = () => {
    setError('')
    setFieldErrors({})
    if (stepIndex > 0) setStep(STEPS[stepIndex - 1])
  }

  const handleChange = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target
    setDraft((prev) => ({ ...prev, [name]: value }))
    if (name in fieldErrors) setFieldErrors((prev) => ({ ...prev, [name]: '' }))
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (step !== 'review') {
      await goNext()
      return
    }

    setLoading(true)
    setError('')
    try {
      if (!createdListingId) {
        throw new Error('Listing was not created yet')
      }
      if (isEditMode) {
        await updateListing(token, createdListingId, {
          title: draft.title.trim(),
          description: draft.description.trim(),
          condition: draft.condition as ListingCondition,
          price: Number(draft.price),
          currency: draft.currency,
          city: draft.city.trim(),
          state: draft.state.trim() || undefined,
          category_id: draft.category_id || undefined
        })
      } else {
        await updateListingStatus(token, createdListingId, { status: 'active' })
      }
      onSuccess()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : `Failed to ${isEditMode ? 'update' : 'create'} listing`)
    } finally {
      setLoading(false)
    }
  }

  const handleCancel = async () => {
    if (hasUnsavedChanges) {
      const confirmed = window.confirm('You have unsaved listing changes. Leave this page?')
      if (!confirmed) return
    }

    // Avoid leaving an empty active listing behind if the user cancels early.
    if (createdListingId && !isEditMode) {
      try {
        await deleteListing(token, createdListingId)
      } catch {
        // Best-effort cleanup.
      }
    }
    onCancel()
  }

  const handleSaveDraft = async () => {
    setLoading(true)
    setError('')
    try {
      if (!createdListingId) {
        throw new Error('Listing was not created yet')
      }
      onSuccess()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to save draft')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="create-listing">
      <h2 className="create-listing-heading">
        <IconAddListing className="create-listing-heading-icon" aria-hidden />
        {isEditMode ? 'Edit Listing' : 'Create Listing'}
      </h2>
      <p className="create-listing-subtitle">Add your item details in a few quick steps.</p>
      <p className="create-listing-step-index">
        Step {stepIndex + 1} of {STEPS.length}
      </p>
      <div className="create-listing-progress" aria-hidden>
        <div className="create-listing-progress-track">
          <div className="create-listing-progress-fill" style={{ width: `${progressPercent}%` }} />
        </div>
      </div>

      <div className="create-listing-steps">
        {STEPS.map((s, i) => (
          <button
            key={s}
            type="button"
            className={`create-listing-step ${step === s ? 'active' : ''} ${i < stepIndex ? 'done' : ''}`}
            onClick={() => i <= stepIndex && setStep(s)}
          >
            {i + 1}. {s.charAt(0).toUpperCase() + s.slice(1)}
          </button>
        ))}
      </div>

      {error && <p className="error-message">{error}</p>}

      <form onSubmit={handleSubmit} className="create-listing-form">
        {step === 'basic' && (
          <div className="create-listing-step-content create-listing-step-content-grid">
            <label className="create-listing-field">
              <span className="create-listing-label">Title *</span>
              <input
                className="create-listing-input create-listing-input-key"
                type="text"
                name="title"
                value={draft.title}
                onChange={handleChange}
                placeholder="e.g. iPhone 12, Wooden Table"
                maxLength={200}
              />
              {fieldErrors.title && (
                <span className="field-error">{fieldErrors.title}</span>
              )}
            </label>

            <label className="create-listing-field">
              <span className="create-listing-label">City *</span>
              <input
                className="create-listing-input create-listing-input-key"
                type="text"
                name="city"
                value={draft.city}
                onChange={handleChange}
                placeholder="Your city"
              />
              {fieldErrors.city && (
                <span className="field-error">{fieldErrors.city}</span>
              )}
            </label>

            <label className="create-listing-field">
              <span className="create-listing-label">Category</span>
              <select
                className="create-listing-input"
                name="category_id"
                value={draft.category_id || ''}
                onChange={handleChange}
              >
                <option value="">Select</option>
                {safeCategories.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
          </div>
        )}

        {step === 'details' && (
          <div className="create-listing-step-content create-listing-step-content-grid">
            <label className="create-listing-field">
              <span className="create-listing-label">Description *</span>
              <textarea
                className="create-listing-input"
                name="description"
                value={draft.description}
                onChange={handleChange}
                placeholder="Describe condition, dimensions, reason for selling..."
                rows={5}
                maxLength={5000}
              />
              <small className="char-count">{draft.description.length}/5000</small>
              {fieldErrors.description && (
                <span className="field-error">{fieldErrors.description}</span>
              )}
            </label>

            <label className="create-listing-field">
              <span className="create-listing-label">Condition *</span>
              <select
                className="create-listing-input"
                name="condition"
                value={draft.condition}
                onChange={handleChange}
              >
                {LISTING_CONDITIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </label>

            <label className="create-listing-field">
              <span className="create-listing-label">Price (INR) *</span>
              <input
                className="create-listing-input"
                type="number"
                name="price"
                value={draft.price || ''}
                onChange={handleChange}
                placeholder="0"
                min={0}
                step={0.01}
              />
              {fieldErrors.price && (
                <span className="field-error">{fieldErrors.price}</span>
              )}
            </label>

            <label className="create-listing-field">
              <span className="create-listing-label">State (optional)</span>
              <input
                className="create-listing-input"
                type="text"
                name="state"
                value={draft.state}
                onChange={handleChange}
                placeholder="State/Region"
              />
            </label>
          </div>
        )}

        {step === 'photos' && (
          <div className="create-listing-step-content">
            <p className="create-listing-photo-hint">
              Add up to 10 photos. First image will be the cover.
            </p>
            <p className="create-listing-photo-subhint">
              Use clear photos in good lighting for faster responses.
            </p>
            <div className="create-listing-photo-card">
              <PhotoUpload
                value={photos}
                onChange={setPhotos}
                maxFiles={10}
                onUpload={async (file) => {
                  if (!createdListingId) {
                    throw new Error('Listing was not created yet')
                  }
                  const res = await uploadListingImageFile(token, createdListingId, file)
                  return res.image.image_url
                }}
              />
            </div>
            <p className="create-listing-photo-skip">
              No photos? You can skip and add them later.
            </p>
          </div>
        )}

        {step === 'review' && (
          <div className="create-listing-step-content create-listing-review">
            <div className="create-listing-review-block">
              <h3>{draft.title || '-'}</h3>
              <p className="create-listing-review-meta">
                📍 {draft.city}{draft.state ? `, ${draft.state}` : ''}
                &nbsp;·&nbsp;
                🏷️ {LISTING_CONDITIONS.find((c) => c.value === draft.condition)?.label ?? draft.condition}
                &nbsp;·&nbsp;
                💰 {draft.currency} {Number(draft.price).toLocaleString()}
              </p>
              {safeCategories.find((c) => c.id === draft.category_id) && (
                <p className="create-listing-review-category">
                  📂 {safeCategories.find((c) => c.id === draft.category_id)?.name}
                </p>
              )}
              <p className="create-listing-review-desc">{draft.description || '-'}</p>
              {photos.length > 0 && (
                <div className="create-listing-review-photos-grid">
                  {photos.map((p, i) => (
                    <img
                      key={p.id}
                      src={p.url || p.preview}
                      alt={`Photo ${i + 1}`}
                      className="create-listing-review-thumb"
                    />
                  ))}
                </div>
              )}
              {photos.length === 0 && (
                <p className="create-listing-review-no-photos">No photos attached</p>
              )}
            </div>
          </div>
        )}

        <div className="button-group create-listing-actions">
          {stepIndex > 0 ? (
            <button type="button" className="profile-edit-btn secondary" onClick={goPrev}>
              Back
            </button>
          ) : (
            <button
              type="button"
              className="profile-edit-btn secondary"
              onClick={() => {
                void handleCancel()
              }}
              disabled={stepLoading}
            >
              Cancel
            </button>
          )}
          {step !== 'review' ? (
            <button
              type="button"
              className="profile-edit-btn primary"
              onClick={() => {
                void goNext()
              }}
              disabled={stepLoading}
            >
              Next
            </button>
          ) : (
            <>
              {!isEditMode && (
                <button
                  type="button"
                  className="profile-edit-btn secondary"
                  onClick={() => {
                    void handleSaveDraft()
                  }}
                  disabled={loading}
                >
                  {loading ? 'Saving...' : 'Save Draft'}
                </button>
              )}
              <button type="submit" className="profile-edit-btn primary" disabled={loading}>
                {loading ? (isEditMode ? 'Saving...' : 'Publishing...') : isEditMode ? 'Save Changes' : 'Publish Listing'}
              </button>
            </>
          )}
        </div>
      </form>
    </div>
  )
}
