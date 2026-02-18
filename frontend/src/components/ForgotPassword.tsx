import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'
import { requestPasswordReset } from '../api/auth'
import { logError, logInfo } from '../utils/logger'
import { isValidEmail } from '../utils/validation'
import { IconBack, IconEmail, IconLock } from './Icons'

interface ForgotPasswordProps {
  onBack: () => void
  onGoToReset: () => void
}

export default function ForgotPassword({ onBack, onGoToReset }: ForgotPasswordProps) {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  function handleEmailChange(event: ChangeEvent<HTMLInputElement>) {
    setEmail(event.target.value)
    setError('')
    setMessage('')
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setLoading(true)
    setError('')
    setMessage('')

    if (!isValidEmail(email)) {
      setError('Please enter a valid email address')
      setLoading(false)
      return
    }

    try {
      const response = await requestPasswordReset(email)
      setMessage(response.message || 'OTP sent to your email. Please check your inbox.')
      logInfo('auth.password_reset.request.success', { email })
    } catch (err: unknown) {
      logError('auth.password_reset.request.failed', {
        email,
        error: err instanceof Error ? err.message : 'unknown error'
      })
      setError(err instanceof Error ? err.message : 'Failed to send reset email')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="forgot-password">
      <h2 className="forgot-password-title">
        <IconLock className="forgot-password-title-icon" aria-hidden />
        <span>Reset Password</span>
      </h2>
      <p className="subtitle forgot-password-subtitle">
        We will send a one-time passcode (OTP) to your email.
      </p>

      {error && <p className="error-message">{error}</p>}
      {message && <p className="success-message">{message}</p>}

      <form onSubmit={handleSubmit} className="auth-form forgot-password-form">
        <div className="auth-form-card">
          <label className="auth-form-label" htmlFor="forgot-password-email">
            <IconEmail className="auth-form-label-icon" aria-hidden />
            <span>Email</span>
          </label>
          <input
            id="forgot-password-email"
            type="email"
            className="auth-form-input"
            value={email}
            onChange={handleEmailChange}
            placeholder="your.email@example.com"
            required
          />
        </div>

        <button type="submit" className="profile-edit-btn primary forgot-password-submit" disabled={loading}>
          {loading ? 'Sending...' : 'Send OTP'}
        </button>

        <div className="forgot-password-actions">
          <button type="button" className="profile-edit-btn secondary" onClick={onGoToReset}>
            Have OTP? Reset now
          </button>
          <button type="button" className="profile-edit-btn secondary" onClick={onBack}>
            <IconBack className="profile-action-icon" aria-hidden />
            <span>Back to Login</span>
          </button>
        </div>
      </form>
    </div>
  )
}
