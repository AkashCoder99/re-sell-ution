import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'
import { requestPasswordReset } from '../api/auth'
import { isValidEmail } from '../utils/validation'

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
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to send reset email')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="forgot-password">
      <h2>🔒 Reset Password</h2>
      <p className="subtitle">We'll send a one-time passcode (OTP) to your email</p>

      {error && <p className="error-message">{error}</p>}
      {message && <p className="success-message">{message}</p>}

      <form onSubmit={handleSubmit} className="auth-form">
        <label>
          Email
          <input
            type="email"
            value={email}
            onChange={handleEmailChange}
            placeholder="your.email@example.com"
            required
          />
        </label>

        <div className="button-group">
          <button type="submit" disabled={loading}>
            {loading ? 'Sending...' : 'Send OTP'}
          </button>
          <button type="button" onClick={onGoToReset}>
            Have OTP? Reset now
          </button>
          <button type="button" className="secondary" onClick={onBack}>
            Back to Login
          </button>
        </div>
      </form>
    </div>
  )
}
