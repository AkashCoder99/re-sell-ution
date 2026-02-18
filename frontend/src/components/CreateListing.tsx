/**
 * F12 - Create listing: multi-step flow (basic info -> details -> photos -> review)
 */

import { useState, useEffect } from 'react'
import type { ChangeEvent, FormEvent } from 'react'
import type { CreateListingDraft, ListingCondition } from '../types/listing'
import { LISTING_CONDITIONS } from '../types/listing'
import {
  validateListingTitle,
  validateListingDescription,
  validateListingPrice
} from '../utils/validation'
import PhotoUpload, { type PhotoItem } from './PhotoUpload'
import type { Category } from '../types/listing'
import { getCategories, createListing } from '../api/listings'
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
}

export default function CreateListing({
  token,
  userCity,
  onSuccess,
  onCancel
}: CreateListingProps) {
  const [step, setStep] = useState<Step>('basic')
  const [draft, setDraft] = useState<CreateListingDraft>(() => ({
    ...defaultDraft,
    city: userCity
  }))
  const [photos, setPhotos] = useState<PhotoItem[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    getCategories(token)
      .then((res) => setCategories(res.categories))
      .catch(() => setCategories([]))
  }, [token])

  const stepIndex = STEPS.indexOf(step)
  const progressPercent = ((stepIndex + 1) / STEPS.length) * 100

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

  const goNext = () => {
    setError('')
    setFieldErrors({})
    if (step === 'basic' && !validateBasic()) return
    if (step === 'details' && !validateDetails()) return
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
      goNext()
      return
    }

    setLoading(true)
    setError('')
    try {
      const image_urls = photos
        .filter((p) => p.url || (p.preview && p.status === 'done'))
        .map((p) => p.url || p.preview)
      await createListing(token, {
        title: draft.title.trim(),
        description: draft.description.trim(),
        condition: draft.condition as ListingCondition,
        price: Number(draft.price),
        currency: draft.currency,
        city: draft.city.trim(),
        state: draft.state.trim() || undefined,
        category_id: draft.category_id || undefined,
        image_urls
      })
      onSuccess()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to create listing')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="create-listing">
      <h2 className="create-listing-heading">
        <IconAddListing className="create-listing-heading-icon" aria-hidden />
        Create Listing
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
                {categories.map((c) => (
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
                  // Mock: use object URL as "uploaded" URL so we can submit with image_urls
                  return Promise.resolve(URL.createObjectURL(file))
                }}
              />
            </div>
          </div>
        )}

        {step === 'review' && (
          <div className="create-listing-step-content create-listing-review">
            <div className="create-listing-review-block">
              <h3>{draft.title || '-'}</h3>
              <p className="create-listing-review-meta">
                {draft.city}
                {draft.state ? `, ${draft.state}` : ''} | {draft.condition} | INR {draft.price}
              </p>
              <p className="create-listing-review-desc">{draft.description || '-'}</p>
              <p className="create-listing-review-photos">
                {photos.length} photo(s) attached
              </p>
            </div>
          </div>
        )}

        <div className="button-group create-listing-actions">
          {stepIndex > 0 ? (
            <button type="button" className="profile-edit-btn secondary" onClick={goPrev}>
              Back
            </button>
          ) : (
            <button type="button" className="profile-edit-btn secondary" onClick={onCancel}>
              Cancel
            </button>
          )}
          {step !== 'review' ? (
            <button type="button" className="profile-edit-btn primary" onClick={goNext}>
              Next
            </button>
          ) : (
            <button type="submit" className="profile-edit-btn primary" disabled={loading}>
              {loading ? 'Publishing...' : 'Publish Listing'}
            </button>
          )}
        </div>
      </form>
    </div>
  )
}
