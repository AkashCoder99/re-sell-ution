import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'

interface CitySelectorProps {
  currentCity?: string
  onCitySelected: (city: string) => void
  onSkip?: () => void
}

export default function CitySelector({ currentCity, onCitySelected, onSkip }: CitySelectorProps) {
  const [selectedCity, setSelectedCity] = useState(currentCity || '')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // Common cities list (can be extended or fetched from backend)
  const popularCities = [
    'New York',
    'Los Angeles',
    'Chicago',
    'Houston',
    'Phoenix',
    'Philadelphia',
    'San Antonio',
    'San Diego',
    'Dallas',
    'San Jose',
    'Austin',
    'Jacksonville',
    'Fort Worth',
    'Columbus',
    'Charlotte',
    'San Francisco',
    'Indianapolis',
    'Seattle',
    'Denver',
    'Boston'
  ]

  function handleCityChange(event: ChangeEvent<HTMLSelectElement | HTMLInputElement>) {
    setSelectedCity(event.target.value)
    setError('')
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!selectedCity.trim()) {
      setError('Please select or enter a city')
      return
    }

    setLoading(true)
    setError('')

    try {
      await onCitySelected(selectedCity.trim())
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to update city')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="city-selector">
      <h2>📍 Choose Your City</h2>
      <p className="subtitle">Select your location to discover local listings</p>

      {error && <p className="error-message">{error}</p>}

      <form onSubmit={handleSubmit} className="city-form">
        <label>
          Select from popular cities
          <select value={selectedCity} onChange={handleCityChange}>
            <option value="">-- Choose a city --</option>
            {popularCities.map((city) => (
              <option key={city} value={city}>
                {city}
              </option>
            ))}
          </select>
        </label>

        <div className="divider">OR</div>

        <label>
          Enter your city manually
          <input
            type="text"
            value={selectedCity}
            onChange={handleCityChange}
            placeholder="e.g., Miami"
          />
        </label>

        <div className="button-group">
          <button type="submit" disabled={loading || !selectedCity.trim()}>
            {loading ? 'Saving...' : 'Confirm City'}
          </button>
          {onSkip && (
            <button type="button" className="secondary" onClick={onSkip}>
              Skip for now
            </button>
          )}
        </div>
      </form>
    </div>
  )
}
