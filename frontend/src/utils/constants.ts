/**
 * Constants used throughout the application
 */

// API Configuration
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
export const USE_MOCK_API = import.meta.env.VITE_USE_MOCK === 'true'

// Authentication
export const TOKEN_KEY = 'auth_token'
export const USER_KEY = 'user_data'

// Validation Limits
export const PASSWORD_MIN_LENGTH = 8
export const NAME_MIN_LENGTH = 2
export const NAME_MAX_LENGTH = 100
export const BIO_MAX_LENGTH = 500
export const CITY_MIN_LENGTH = 2
export const CITY_MAX_LENGTH = 100

// Popular Cities
export const POPULAR_CITIES = [
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
] as const

// UI Messages
export const MESSAGES = {
  LOGIN_SUCCESS: 'Login successful!',
  LOGOUT_SUCCESS: 'Logged out successfully',
  REGISTER_SUCCESS: 'Registration successful! Please log in.',
  PROFILE_UPDATE_SUCCESS: 'Profile updated successfully!',
  CITY_UPDATE_SUCCESS: 'City updated successfully!',
  PASSWORD_RESET_SENT: 'Password reset link sent to your email',
  ERROR_GENERIC: 'Something went wrong. Please try again.',
  ERROR_NETWORK: 'Network error. Please check your connection.',
  ERROR_AUTH: 'Authentication failed. Please try again.',
} as const
