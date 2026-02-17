import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'
import type { PublicUser, UpdateProfileRequest } from '../types/user'
import { IconEdit, IconUser, IconLocation, IconBio, IconImage } from './Icons'

interface ProfileEditProps {
  user: PublicUser
  onUpdate: (updates: UpdateProfileRequest) => Promise<void>
  onCancel: () => void
}

export default function ProfileEdit({ user, onUpdate, onCancel }: ProfileEditProps) {
  const [formData, setFormData] = useState<UpdateProfileRequest>({
    full_name: user.full_name,
    city: user.city || '',
    bio: user.bio || '',
    photo_url: user.photo_url || ''
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  function handleChange(event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) {
    const { name, value } = event.target
    setFormData((prev) => ({ ...prev, [name]: value }))
    setError('')
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setLoading(true)
    setError('')

    try {
      await onUpdate(formData)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to update profile')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="profile-edit">
      <h2 className="profile-edit-title">
        <IconEdit className="profile-edit-title-icon" aria-hidden />
        Edit Profile
      </h2>

      {error && <p className="error-message">{error}</p>}

      <form onSubmit={handleSubmit} className="profile-edit-form">
        <div className="profile-edit-card">
          <label className="profile-edit-label" htmlFor="profile-full_name">
            <IconUser className="profile-edit-label-icon" aria-hidden />
            <span>Full Name</span>
          </label>
          <input
            id="profile-full_name"
            type="text"
            name="full_name"
            value={formData.full_name}
            onChange={handleChange}
            className="profile-edit-input"
            required
            aria-required
          />
        </div>

        <div className="profile-edit-card">
          <label className="profile-edit-label" htmlFor="profile-city">
            <IconLocation className="profile-edit-label-icon" aria-hidden />
            <span>City</span>
          </label>
          <input
            id="profile-city"
            type="text"
            name="city"
            value={formData.city}
            onChange={handleChange}
            className="profile-edit-input"
            placeholder="e.g., New York"
          />
        </div>

        <div className="profile-edit-card">
          <label className="profile-edit-label" htmlFor="profile-bio">
            <IconBio className="profile-edit-label-icon" aria-hidden />
            <span>Bio</span>
          </label>
          <textarea
            id="profile-bio"
            name="bio"
            value={formData.bio}
            onChange={handleChange}
            className="profile-edit-textarea"
            placeholder="Tell us about yourself"
            rows={4}
            maxLength={500}
            aria-describedby="profile-bio-count"
          />
          <span id="profile-bio-count" className="profile-edit-char-count">
            {formData.bio.length}/500 characters
          </span>
        </div>

        <div className="profile-edit-card">
          <label className="profile-edit-label" htmlFor="profile-photo_url">
            <IconImage className="profile-edit-label-icon" aria-hidden />
            <span>Profile Photo URL</span>
          </label>
          <input
            id="profile-photo_url"
            type="url"
            name="photo_url"
            value={formData.photo_url}
            onChange={handleChange}
            className="profile-edit-input"
            placeholder="https://example.com/photo.jpg"
          />
        </div>

        <div className="profile-edit-actions">
          <button
            type="submit"
            className="profile-edit-btn primary"
            disabled={loading}
          >
            {loading ? 'Saving...' : 'Save Changes'}
          </button>
          <button
            type="button"
            className="profile-edit-btn secondary"
            onClick={onCancel}
            disabled={loading}
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}
