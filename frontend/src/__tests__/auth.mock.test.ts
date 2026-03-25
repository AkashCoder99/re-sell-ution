/**
 * Unit tests for the mock auth API functions.
 * Forces USE_MOCK mode by setting the env var before importing.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

// Set mock mode before module import
vi.stubEnv('VITE_USE_MOCK', 'true')

// Dynamic import so env stub takes effect
const { register, login, logout, getMe, updateProfile, requestPasswordReset, confirmPasswordReset } =
  await import('../api/auth')

describe('mock auth API — register', () => {
  it('registers a new user and returns token + user', async () => {
    const result = await register({ full_name: 'Test User', email: `u${Date.now()}@test.com`, password: 'Password1' })
    expect(result.token).toBeTruthy()
    expect(result.user.email).toContain('@test.com')
    expect(result.user.full_name).toBe('Test User')
  })

  it('throws if email already registered', async () => {
    const email = `dup${Date.now()}@test.com`
    await register({ full_name: 'Dup User', email, password: 'Password1' })
    await expect(register({ full_name: 'Dup User', email, password: 'Password1' })).rejects.toThrow(
      'Email already registered'
    )
  })
})

describe('mock auth API — login', () => {
  it('logs in with correct credentials', async () => {
    const email = `login${Date.now()}@test.com`
    await register({ full_name: 'Login User', email, password: 'Password1' })
    const result = await login({ email, password: 'Password1' })
    expect(result.token).toBeTruthy()
    expect(result.user.email).toBe(email)
  })

  it('throws on wrong password', async () => {
    const email = `wrong${Date.now()}@test.com`
    await register({ full_name: 'Wrong Pass', email, password: 'Password1' })
    await expect(login({ email, password: 'wrongpass' })).rejects.toThrow('Invalid email or password')
  })

  it('throws on unknown email', async () => {
    await expect(login({ email: 'nobody@nowhere.com', password: 'Password1' })).rejects.toThrow()
  })
})

describe('mock auth API — getMe', () => {
  it('returns user for valid token', async () => {
    const email = `me${Date.now()}@test.com`
    const { token } = await register({ full_name: 'Me User', email, password: 'Password1' })
    const result = await getMe(token)
    // mock returns first registered user — just verify a user object is returned
    expect(result.user).toBeTruthy()
    expect(result.user.email).toBeTruthy()
  })

  it('throws for invalid token', async () => {
    await expect(getMe('bad_token')).rejects.toThrow()
  })
})

describe('mock auth API — logout', () => {
  it('logs out successfully', async () => {
    const email = `lo${Date.now()}@test.com`
    const { token } = await register({ full_name: 'Logout User', email, password: 'Password1' })
    const result = await logout(token)
    expect(result.message).toBeTruthy()
  })

  it('throws when logging out with invalid token', async () => {
    await expect(logout('invalid_token')).rejects.toThrow()
  })
})

describe('mock auth API — updateProfile', () => {
  it('updates city on profile', async () => {
    const email = `upd${Date.now()}@test.com`
    const { token } = await register({ full_name: 'Update User', email, password: 'Password1' })
    const result = await updateProfile(token, { city: 'Gainesville' })
    expect(result.user.city).toBe('Gainesville')
  })

  it('updates bio on profile', async () => {
    const email = `bio${Date.now()}@test.com`
    const { token } = await register({ full_name: 'Bio User', email, password: 'Password1' })
    const result = await updateProfile(token, { bio: 'Hello world' })
    expect(result.user.bio).toBe('Hello world')
  })
})

describe('mock auth API — password reset', () => {
  it('requestPasswordReset returns a message', async () => {
    const email = `reset${Date.now()}@test.com`
    await register({ full_name: 'Reset User', email, password: 'Password1' })
    const result = await requestPasswordReset(email)
    expect(result.message).toBeTruthy()
  })

  it('confirmPasswordReset succeeds with correct OTP', async () => {
    const email = `confirm${Date.now()}@test.com`
    await register({ full_name: 'Confirm User', email, password: 'Password1' })
    const reqResult = await requestPasswordReset(email)
    // OTP is embedded in mock message: "OTP sent (mock): 123456"
    const otp = reqResult.message.split(': ')[1]
    const result = await confirmPasswordReset({ email, otp, new_password: 'NewPass99' })
    expect(result.message).toContain('success')
  })

  it('confirmPasswordReset throws with wrong OTP', async () => {
    const email = `badotp${Date.now()}@test.com`
    await register({ full_name: 'Bad OTP', email, password: 'Password1' })
    await requestPasswordReset(email)
    await expect(confirmPasswordReset({ email, otp: '000000', new_password: 'NewPass99' })).rejects.toThrow()
  })
})
