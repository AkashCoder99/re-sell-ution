import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'
import { confirmPasswordReset } from '../api/auth'
import { logError, logInfo } from '../utils/logger'
import { isValidEmail } from '../utils/validation'

interface ResetPasswordProps {
  onBack: () => void
}

export default function ResetPassword({ onBack }: ResetPasswordProps) {
  const [email, setEmail] = useState('')
  const [otp, setOtp] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  function onEmailChange(event: ChangeEvent<HTMLInputElement>) {
    setEmail(event.target.value)
    setError('')
    setMessage('')
  }

  function onOtpChange(event: ChangeEvent<HTMLInputElement>) {
    setOtp(event.target.value)
    setError('')
    setMessage('')
  }

  function onNewPasswordChange(event: ChangeEvent<HTMLInputElement>) {
    setNewPassword(event.target.value)
    setError('')
    setMessage('')
  }

  function onConfirmPasswordChange(event: ChangeEvent<HTMLInputElement>) {
    setConfirmPassword(event.target.value)
    setError('')
    setMessage('')
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setLoading(true)
    setError('')
    setMessage('')

    if (!isValidEmail(email)) {
      setError('Valid email is required')
      setLoading(false)
      return
    }

    if (!otp.trim()) {
      setError('OTP is required')
      setLoading(false)
      return
    }

    if (newPassword.length < 8) {
      setError('New password must be at least 8 characters')
      setLoading(false)
      return
    }

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match')
      setLoading(false)
      return
    }

    try {
      const response = await confirmPasswordReset({
        email: email.trim(),
        otp: otp.trim(),
        new_password: newPassword
      })
      setMessage(response.message || 'Password reset successful. You can now log in.')
      logInfo('auth.password_reset.confirm.success', { email: email.trim() })
      setEmail('')
      setOtp('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err: unknown) {
      logError('auth.password_reset.confirm.failed', {
        email: email.trim(),
        error: err instanceof Error ? err.message : 'unknown error'
      })
      setError(err instanceof Error ? err.message : 'Failed to reset password')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="forgot-password">
      <h2>🔐 Set New Password</h2>
      <p className="subtitle">Enter your email, OTP, and choose a new password</p>

      {error && <p className="error-message">{error}</p>}
      {message && <p className="success-message">{message}</p>}

      <form onSubmit={handleSubmit} className="auth-form">
        <label>
          Email
          <input type="email" value={email} onChange={onEmailChange} placeholder="your.email@example.com" required />
        </label>

        <label>
          OTP
          <input type="text" value={otp} onChange={onOtpChange} placeholder="Enter 6-digit OTP" required />
        </label>

        <label>
          New password
          <input type="password" value={newPassword} onChange={onNewPasswordChange} minLength={8} required />
        </label>

        <label>
          Confirm new password
          <input type="password" value={confirmPassword} onChange={onConfirmPasswordChange} minLength={8} required />
        </label>

        <div className="button-group">
          <button type="submit" disabled={loading}>
            {loading ? 'Resetting...' : 'Reset Password'}
          </button>
          <button type="button" className="secondary" onClick={onBack}>
            Back
          </button>
        </div>
      </form>
    </div>
  )
}
