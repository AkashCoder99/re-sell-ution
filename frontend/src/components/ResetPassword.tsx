import type { ChangeEvent, FormEvent } from 'react'
import { useState } from 'react'
import { confirmPasswordReset } from '../api/auth'

interface ResetPasswordProps {
  onBack: () => void
}

export default function ResetPassword({ onBack }: ResetPasswordProps) {
  const [token, setToken] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  function onTokenChange(event: ChangeEvent<HTMLInputElement>) {
    setToken(event.target.value)
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

    if (!token.trim()) {
      setError('Reset token is required')
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
        token: token.trim(),
        new_password: newPassword
      })
      setMessage(response.message || 'Password reset successful. You can now log in.')
      setToken('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to reset password')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="forgot-password">
      <h2>🔐 Set New Password</h2>
      <p className="subtitle">Enter your reset token and choose a new password</p>

      {error && <p className="error-message">{error}</p>}
      {message && <p className="success-message">{message}</p>}

      <form onSubmit={handleSubmit} className="auth-form">
        <label>
          Reset token
          <input type="text" value={token} onChange={onTokenChange} placeholder="Paste token here" required />
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
