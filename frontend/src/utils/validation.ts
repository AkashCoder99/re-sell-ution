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
