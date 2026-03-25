import { describe, it, expect } from 'vitest'
import {
  isValidEmail,
  isValidPassword,
  isStrongPassword,
  validateFullName,
  validateBio,
  validateCity,
  validateListingTitle,
  validateListingDescription,
  validateListingPrice
} from '../utils/validation'

describe('isValidEmail', () => {
  it('accepts a valid email', () => {
    expect(isValidEmail('user@example.com')).toBe(true)
  })
  it('rejects email without @', () => {
    expect(isValidEmail('userexample.com')).toBe(false)
  })
  it('rejects email without domain', () => {
    expect(isValidEmail('user@')).toBe(false)
  })
  it('rejects empty string', () => {
    expect(isValidEmail('')).toBe(false)
  })
})

describe('isValidPassword', () => {
  it('accepts password with 8+ characters', () => {
    expect(isValidPassword('abcdefgh')).toBe(true)
  })
  it('rejects password shorter than 8 characters', () => {
    expect(isValidPassword('abc')).toBe(false)
  })
  it('accepts exactly 8 characters', () => {
    expect(isValidPassword('12345678')).toBe(true)
  })
})

describe('isStrongPassword', () => {
  it('accepts password with upper, lower, and digit', () => {
    expect(isStrongPassword('Abcdef1g')).toBe(true)
  })
  it('rejects password without uppercase', () => {
    expect(isStrongPassword('abcdef1g')).toBe(false)
  })
  it('rejects password without digit', () => {
    expect(isStrongPassword('Abcdefgh')).toBe(false)
  })
  it('rejects short password', () => {
    expect(isStrongPassword('Ab1')).toBe(false)
  })
})

describe('validateFullName', () => {
  it('returns null for valid name', () => {
    expect(validateFullName('Krishna')).toBeNull()
  })
  it('returns error for empty name', () => {
    expect(validateFullName('')).not.toBeNull()
  })
  it('returns error for single character', () => {
    expect(validateFullName('K')).not.toBeNull()
  })
  it('returns error for name over 100 chars', () => {
    expect(validateFullName('a'.repeat(101))).not.toBeNull()
  })
  it('trims whitespace before validating', () => {
    expect(validateFullName('  ')).not.toBeNull()
  })
})

describe('validateBio', () => {
  it('returns null for empty bio', () => {
    expect(validateBio('')).toBeNull()
  })
  it('returns null for bio under 500 chars', () => {
    expect(validateBio('Hello world')).toBeNull()
  })
  it('returns error for bio over 500 chars', () => {
    expect(validateBio('a'.repeat(501))).not.toBeNull()
  })
})

describe('validateCity', () => {
  it('returns null for valid city', () => {
    expect(validateCity('Gainesville')).toBeNull()
  })
  it('returns error for empty city', () => {
    expect(validateCity('')).not.toBeNull()
  })
  it('returns error for single char city', () => {
    expect(validateCity('A')).not.toBeNull()
  })
  it('returns error for city over 100 chars', () => {
    expect(validateCity('a'.repeat(101))).not.toBeNull()
  })
})

describe('validateListingTitle', () => {
  it('returns null for valid title', () => {
    expect(validateListingTitle('iPhone 13')).toBeNull()
  })
  it('returns error for empty title', () => {
    expect(validateListingTitle('')).not.toBeNull()
  })
  it('returns error for title under 3 chars', () => {
    expect(validateListingTitle('ab')).not.toBeNull()
  })
  it('returns error for title over 200 chars', () => {
    expect(validateListingTitle('a'.repeat(201))).not.toBeNull()
  })
})

describe('validateListingDescription', () => {
  it('returns null for valid description', () => {
    expect(validateListingDescription('This is a great item in good condition.')).toBeNull()
  })
  it('returns error for empty description', () => {
    expect(validateListingDescription('')).not.toBeNull()
  })
  it('returns error for description under 10 chars', () => {
    expect(validateListingDescription('Short')).not.toBeNull()
  })
  it('returns error for description over 5000 chars', () => {
    expect(validateListingDescription('a'.repeat(5001))).not.toBeNull()
  })
})

describe('validateListingPrice', () => {
  it('returns null for valid price', () => {
    expect(validateListingPrice(100)).toBeNull()
  })
  it('returns null for zero price', () => {
    expect(validateListingPrice(0)).toBeNull()
  })
  it('returns error for negative price', () => {
    expect(validateListingPrice(-1)).not.toBeNull()
  })
  it('returns error for non-numeric value', () => {
    expect(validateListingPrice('abc')).not.toBeNull()
  })
  it('returns error for price exceeding max', () => {
    expect(validateListingPrice(1_000_000_000)).not.toBeNull()
  })
})
