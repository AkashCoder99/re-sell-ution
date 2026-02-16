import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'
import { POPULAR_CITIES } from '../utils/constants'
import { IconLocation } from './Icons'

interface CitySelectorProps {
  currentCity?: string
  onCitySelected: (city: string) => void
  onSkip?: () => void
}

export default function CitySelector({ currentCity, onCitySelected, onSkip }: CitySelectorProps) {
  const [selectedCity, setSelectedCity] = useState(currentCity || '')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [searchQuery, setSearchQuery] = useState('')

  // Use imported popular cities
  const popularCities = [...POPULAR_CITIES]

  function handleCityChange(event: ChangeEvent<HTMLSelectElement | HTMLInputElement>) {
    setSelectedCity(event.target.value)
    setError('')
  }

  const filteredCities = popularCities.filter((city) =>
    city.toLowerCase().includes(searchQuery.toLowerCase())
  )

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
      <h2 className="city-selector-title">
        <IconLocation className="city-selector-title-icon" aria-hidden />
        Choose Your City
      </h2>
      <p className="subtitle">Select your location to discover local listings</p>

      {error && <p className="error-message">{error}</p>}

      <form onSubmit={handleSubmit} className="city-form">
        <label>
          Search cities
          <input
            type="text"
            placeholder="Type to search..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ marginBottom: '8px' }}
          />
        </label>

        <label>
          Select from popular cities
          <select value={selectedCity} onChange={handleCityChange}>
            <option value="">-- Choose a city --</option>
            {filteredCities.map((city) => (
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
