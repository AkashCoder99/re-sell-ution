import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'
import type { PublicUser, UpdateProfileRequest } from '../types/user'

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
      <h2>✏️ Edit Profile</h2>

      {error && <p className="error-message">{error}</p>}

      <form onSubmit={handleSubmit} className="profile-form">
        <label>
          Full Name
          <input
            type="text"
            name="full_name"
            value={formData.full_name}
            onChange={handleChange}
            required
          />
        </label>

        <label>
          City
          <input
            type="text"
            name="city"
            value={formData.city}
            onChange={handleChange}
            placeholder="e.g., New York"
          />
        </label>

        <label>
          Bio
          <textarea
            name="bio"
            value={formData.bio}
            onChange={handleChange}
            placeholder="Tell us about yourself"
            rows={4}
          />
        </label>

        <label>
          Profile Photo URL
          <input
            type="url"
            name="photo_url"
            value={formData.photo_url}
            onChange={handleChange}
            placeholder="https://example.com/photo.jpg"
          />
        </label>

        <div className="button-group">
          <button type="submit" disabled={loading}>
            {loading ? 'Saving...' : 'Save Changes'}
          </button>
          <button type="button" className="secondary" onClick={onCancel} disabled={loading}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}
