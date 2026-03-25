import { describe, it, expect } from 'vitest'
import {
  API_BASE_URL,
  TOKEN_KEY,
  USER_KEY,
  PASSWORD_MIN_LENGTH,
  NAME_MIN_LENGTH,
  NAME_MAX_LENGTH,
  BIO_MAX_LENGTH,
  CITY_MIN_LENGTH,
  CITY_MAX_LENGTH,
  POPULAR_CITIES,
  MESSAGES
} from '../utils/constants'

describe('constants', () => {
  it('API_BASE_URL defaults to localhost', () => {
    expect(API_BASE_URL).toBe('http://localhost:8080')
  })

  it('TOKEN_KEY is auth_token', () => {
    expect(TOKEN_KEY).toBe('auth_token')
  })

  it('USER_KEY is user_data', () => {
    expect(USER_KEY).toBe('user_data')
  })

  it('PASSWORD_MIN_LENGTH is 8', () => {
    expect(PASSWORD_MIN_LENGTH).toBe(8)
  })

  it('NAME_MIN_LENGTH is 2', () => {
    expect(NAME_MIN_LENGTH).toBe(2)
  })

  it('NAME_MAX_LENGTH is 100', () => {
    expect(NAME_MAX_LENGTH).toBe(100)
  })

  it('BIO_MAX_LENGTH is 500', () => {
    expect(BIO_MAX_LENGTH).toBe(500)
  })

  it('CITY_MIN_LENGTH is 2', () => {
    expect(CITY_MIN_LENGTH).toBe(2)
  })

  it('CITY_MAX_LENGTH is 100', () => {
    expect(CITY_MAX_LENGTH).toBe(100)
  })

  it('POPULAR_CITIES contains at least 10 cities', () => {
    expect(POPULAR_CITIES.length).toBeGreaterThanOrEqual(10)
  })

  it('POPULAR_CITIES contains New York', () => {
    expect(POPULAR_CITIES).toContain('New York')
  })

  it('MESSAGES.ERROR_GENERIC is defined', () => {
    expect(MESSAGES.ERROR_GENERIC).toBeTruthy()
  })

  it('MESSAGES.LOGIN_SUCCESS is defined', () => {
    expect(MESSAGES.LOGIN_SUCCESS).toBeTruthy()
  })

  it('MESSAGES.LOGOUT_SUCCESS is defined', () => {
    expect(MESSAGES.LOGOUT_SUCCESS).toBeTruthy()
  })
})
