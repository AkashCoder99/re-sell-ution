/**
 * Utility functions for form validation
 */

export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

export function isValidPassword(password: string): boolean {
  // At least 8 characters
  return password.length >= 8
}

export function isStrongPassword(password: string): boolean {
  // At least 8 characters, 1 uppercase, 1 lowercase, 1 number
  const strongPasswordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$/
  return strongPasswordRegex.test(password)
}

export function validateFullName(name: string): string | null {
  const trimmed = name.trim()
  if (!trimmed) {
    return 'Name is required'
  }
  if (trimmed.length < 2) {
    return 'Name must be at least 2 characters'
  }
  if (trimmed.length > 100) {
    return 'Name must be less than 100 characters'
  }
  return null
}

export function validateBio(bio: string): string | null {
  if (bio.length > 500) {
    return 'Bio must be less than 500 characters'
  }
  return null
}

export function validateCity(city: string): string | null {
  const trimmed = city.trim()
  if (!trimmed) {
    return 'City is required'
  }
  if (trimmed.length < 2) {
    return 'City name must be at least 2 characters'
  }
  if (trimmed.length > 100) {
    return 'City name must be less than 100 characters'
  }
  return null
}

// Listing validation (EPIC-03)
export function validateListingTitle(title: string): string | null {
  const trimmed = title.trim()
  if (!trimmed) return 'Title is required'
  if (trimmed.length < 3) return 'Title must be at least 3 characters'
  if (trimmed.length > 200) return 'Title must be less than 200 characters'
  return null
}

export function validateListingDescription(description: string): string | null {
  const trimmed = description.trim()
  if (!trimmed) return 'Description is required'
  if (trimmed.length < 10) return 'Description must be at least 10 characters'
  if (trimmed.length > 5000) return 'Description must be less than 5000 characters'
  return null
}

export function validateListingPrice(price: unknown): string | null {
  const n = Number(price)
  if (Number.isNaN(n)) return 'Please enter a valid price'
  if (n < 0) return 'Price cannot be negative'
  if (n > 999_999_999.99) return 'Price is too high'
  return null
}
