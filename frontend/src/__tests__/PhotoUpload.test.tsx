import { describe, it, expect } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { useState } from 'react'
import PhotoUpload, { type PhotoItem } from '../components/PhotoUpload'

function Wrapper() {
  const [items, setItems] = useState<PhotoItem[]>([])
  return <PhotoUpload value={items} onChange={setItems} />
}

describe('PhotoUpload', () => {
  it('adds files and shows preview items', () => {
    render(<Wrapper />)
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    const file = new File(['hello'], 'test.png', { type: 'image/png' })
    fireEvent.change(input, { target: { files: [file] } })
    expect(screen.getByAltText('Upload 1')).toBeInTheDocument()
  })
})
