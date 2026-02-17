import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'
import { POPULAR_CITIES } from '../utils/constants'
import { IconLocation, IconSearch } from './Icons'

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
      <p className="city-selector-subtitle">Select your location to discover local listings</p>

      {error && <p className="error-message">{error}</p>}

      <form onSubmit={handleSubmit} className="city-form">
        <div className="city-selector-card">
          <label className="city-selector-label">
            <IconSearch className="city-selector-label-icon" aria-hidden />
            <span>Search cities</span>
          </label>
          <input
            type="text"
            className="city-selector-search-input"
            placeholder="Type to search..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            aria-label="Search cities"
          />
        </div>

        <div className="city-selector-card">
          <label className="city-selector-label">Select from popular cities</label>
          <select
            value={selectedCity}
            onChange={handleCityChange}
            className="city-selector-select"
            aria-label="Choose a city"
          >
            <option value="">— Choose a city —</option>
            {filteredCities.map((city) => (
              <option key={city} value={city}>
                {city}
              </option>
            ))}
          </select>
        </div>

        <div className="city-selector-divider">
          <span>or</span>
        </div>

        <div className="city-selector-card">
          <label className="city-selector-label">
            <IconLocation className="city-selector-label-icon" aria-hidden />
            <span>Enter your city manually</span>
          </label>
          <input
            type="text"
            className="city-selector-input"
            value={selectedCity}
            onChange={handleCityChange}
            placeholder="e.g., Miami, Austin"
            aria-label="City name"
          />
        </div>

        <div className="city-selector-actions">
          <button
            type="submit"
            className="city-selector-btn primary"
            disabled={loading || !selectedCity.trim()}
          >
            {loading ? 'Saving...' : 'Confirm City'}
          </button>
          {onSkip && (
            <button type="button" className="city-selector-btn secondary" onClick={onSkip}>
              Skip for now
            </button>
          )}
        </div>
      </form>
    </div>
  )
}
