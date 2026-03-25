import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import CitySelector from '../components/CitySelector'

describe('CitySelector', () => {
  it('renders the title', () => {
    render(<CitySelector onCitySelected={vi.fn()} />)
    expect(screen.getByText(/Choose Your City/i)).toBeInTheDocument()
  })

  it('renders the city input', () => {
    render(<CitySelector onCitySelected={vi.fn()} />)
    expect(screen.getByLabelText('City name')).toBeInTheDocument()
  })

  it('renders the city dropdown', () => {
    render(<CitySelector onCitySelected={vi.fn()} />)
    expect(screen.getByLabelText('Choose a city')).toBeInTheDocument()
  })

  it('shows Skip button when onSkip is provided', () => {
    render(<CitySelector onCitySelected={vi.fn()} onSkip={vi.fn()} />)
    expect(screen.getByText('Skip for now')).toBeInTheDocument()
  })

  it('does not show Skip button when onSkip is not provided', () => {
    render(<CitySelector onCitySelected={vi.fn()} />)
    expect(screen.queryByText('Skip for now')).not.toBeInTheDocument()
  })

  it('calls onSkip when Skip button is clicked', () => {
    const onSkip = vi.fn()
    render(<CitySelector onCitySelected={vi.fn()} onSkip={onSkip} />)
    fireEvent.click(screen.getByText('Skip for now'))
    expect(onSkip).toHaveBeenCalledTimes(1)
  })

  it('shows error when submitting without a city', async () => {
    render(<CitySelector onCitySelected={vi.fn()} />)
    fireEvent.submit(screen.getByRole('button', { name: 'Confirm City' }).closest('form')!)
    await waitFor(() => {
      expect(screen.getByText(/Please select or enter a city/i)).toBeInTheDocument()
    })
  })

  it('calls onCitySelected with typed city value', async () => {
    const onCitySelected = vi.fn().mockResolvedValue(undefined)
    render(<CitySelector onCitySelected={onCitySelected} />)
    fireEvent.change(screen.getByLabelText('City name'), { target: { value: 'Gainesville' } })
    fireEvent.click(screen.getByText('Confirm City'))
    await waitFor(() => {
      expect(onCitySelected).toHaveBeenCalledWith('Gainesville')
    })
  })

  it('pre-fills city input with currentCity prop', () => {
    render(<CitySelector currentCity="Austin" onCitySelected={vi.fn()} />)
    expect(screen.getByLabelText('City name')).toHaveValue('Austin')
  })

  it('filters dropdown cities based on search query', () => {
    render(<CitySelector onCitySelected={vi.fn()} />)
    fireEvent.change(screen.getByLabelText('Search cities'), { target: { value: 'San' } })
    expect(screen.getByText('San Antonio')).toBeInTheDocument()
    expect(screen.queryByText('Boston')).not.toBeInTheDocument()
  })

  it('Confirm City button is disabled when city is empty', () => {
    render(<CitySelector onCitySelected={vi.fn()} />)
    expect(screen.getByText('Confirm City')).toBeDisabled()
  })

  it('Confirm City button is enabled when city is entered', () => {
    render(<CitySelector onCitySelected={vi.fn()} />)
    fireEvent.change(screen.getByLabelText('City name'), { target: { value: 'Miami' } })
    expect(screen.getByText('Confirm City')).not.toBeDisabled()
  })
})
