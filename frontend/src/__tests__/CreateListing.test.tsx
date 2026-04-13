import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'

vi.mock('../api/listings', () => ({
  getCategories: vi.fn().mockResolvedValue({ categories: [] }),
  createListing: vi.fn().mockResolvedValue({ listing: { id: 'l1' } })
}))

import CreateListing from '../components/CreateListing'

beforeEach(() => {
  vi.clearAllMocks()
})

describe('CreateListing', () => {
  it('renders the multi-step header', () => {
    render(<CreateListing token="t" userCity="New York" onSuccess={vi.fn()} onCancel={vi.fn()} />)
    expect(screen.getByText('Create Listing')).toBeInTheDocument()
  })

  it('validates required fields before advancing', async () => {
    render(<CreateListing token="t" userCity="" onSuccess={vi.fn()} onCancel={vi.fn()} />)
    fireEvent.click(screen.getByText('Next'))
    await waitFor(() => {
      expect(screen.getByText('City is required')).toBeInTheDocument()
    })
  })

  it('submits after review step', async () => {
    const onSuccess = vi.fn()
    render(<CreateListing token="t" userCity="New York" onSuccess={onSuccess} onCancel={vi.fn()} />)

    fireEvent.change(screen.getByPlaceholderText('e.g. iPhone 12, Wooden Table'), {
      target: { value: 'Lamp' }
    })
    fireEvent.click(screen.getByText('Next'))

    fireEvent.change(screen.getByPlaceholderText('Describe condition, dimensions, reason for selling...'), {
      target: { value: 'Nice lamp' }
    })
    fireEvent.change(screen.getByPlaceholderText('0'), { target: { value: '10' } })
    fireEvent.click(screen.getByText('Next'))

    fireEvent.click(screen.getByText('Next'))
    fireEvent.click(screen.getByText('Publish Listing'))

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalled()
    })
  })
})
