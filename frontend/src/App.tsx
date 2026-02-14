import { useEffect, useMemo, useState } from 'react'
import type { ChangeEvent, FormEvent } from 'react'
import { getMe, login, register } from './api/auth'
import type { LoginRequest, PublicUser, RegisterRequest } from './api/auth'

const defaultLoginForm: LoginRequest = { email: '', password: '' }
const defaultRegisterForm: RegisterRequest = { full_name: '', email: '', password: '' }

export default function App() {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [loginForm, setLoginForm] = useState<LoginRequest>(defaultLoginForm)
  const [registerForm, setRegisterForm] = useState<RegisterRequest>(defaultRegisterForm)
  const [token, setToken] = useState<string>(() => localStorage.getItem('auth_token') || '')
  const [user, setUser] = useState<PublicUser | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [message, setMessage] = useState<string>('')
  const isAuthenticated = useMemo(() => Boolean(token && user), [token, user])

  useEffect(() => {
    if (!token) {
      setUser(null)
      return
    }

    getMe(token)
      .then((data) => {
        setUser(data.user)
      })
      .catch(() => {
        localStorage.removeItem('auth_token')
        setToken('')
        setUser(null)
      })
  }, [token])

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setLoading(true)
    setMessage('')

    try {
      const data = await login(loginForm)
      localStorage.setItem('auth_token', data.token)
      setToken(data.token)
      setUser(data.user)
      setMessage('Login successful.')
      setLoginForm(defaultLoginForm)
    } catch (error: unknown) {
      setMessage(error instanceof Error ? error.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  async function handleRegister(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setLoading(true)
    setMessage('')

    try {
      const data = await register(registerForm)
      localStorage.setItem('auth_token', data.token)
      setToken(data.token)
      setUser(data.user)
      setMessage('Account created and logged in.')
      setRegisterForm(defaultRegisterForm)
    } catch (error: unknown) {
      setMessage(error instanceof Error ? error.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  function handleLogout() {
    localStorage.removeItem('auth_token')
    setToken('')
    setUser(null)
    setMessage('Logged out.')
  }

  function onLoginEmailChange(event: ChangeEvent<HTMLInputElement>) {
    setLoginForm((prev) => ({ ...prev, email: event.target.value }))
  }

  function onLoginPasswordChange(event: ChangeEvent<HTMLInputElement>) {
    setLoginForm((prev) => ({ ...prev, password: event.target.value }))
  }

  function onRegisterNameChange(event: ChangeEvent<HTMLInputElement>) {
    setRegisterForm((prev) => ({ ...prev, full_name: event.target.value }))
  }

  function onRegisterEmailChange(event: ChangeEvent<HTMLInputElement>) {
    setRegisterForm((prev) => ({ ...prev, email: event.target.value }))
  }

  function onRegisterPasswordChange(event: ChangeEvent<HTMLInputElement>) {
    setRegisterForm((prev) => ({ ...prev, password: event.target.value }))
  }

  return (
    <main className="auth-page">
      <section className="auth-card">
        <h1>ReSellution</h1>
        <p className="subtitle">Local marketplace login</p>

        {message && <p className="message">{message}</p>}

        {isAuthenticated ? (
          <div className="profile">
            <h2>Welcome, {user?.full_name}</h2>
            <p>{user?.email}</p>
            <button type="button" onClick={handleLogout}>
              Logout
            </button>
          </div>
        ) : (
          <>
            <div className="mode-toggle">
              <button
                type="button"
                className={mode === 'login' ? 'active' : ''}
                onClick={() => setMode('login')}
              >
                Login
              </button>
              <button
                type="button"
                className={mode === 'register' ? 'active' : ''}
                onClick={() => setMode('register')}
              >
                Register
              </button>
            </div>

            {mode === 'login' ? (
              <form onSubmit={handleLogin} className="auth-form">
                <label>
                  Email
                  <input type="email" value={loginForm.email} onChange={onLoginEmailChange} required />
                </label>

                <label>
                  Password
                  <input type="password" value={loginForm.password} onChange={onLoginPasswordChange} required />
                </label>

                <button type="submit" disabled={loading}>
                  {loading ? 'Please wait...' : 'Login'}
                </button>
              </form>
            ) : (
              <form onSubmit={handleRegister} className="auth-form">
                <label>
                  Full name
                  <input type="text" value={registerForm.full_name} onChange={onRegisterNameChange} required />
                </label>

                <label>
                  Email
                  <input type="email" value={registerForm.email} onChange={onRegisterEmailChange} required />
                </label>

                <label>
                  Password (min 8 chars)
                  <input
                    type="password"
                    value={registerForm.password}
                    onChange={onRegisterPasswordChange}
                    required
                    minLength={8}
                  />
                </label>

                <button type="submit" disabled={loading}>
                  {loading ? 'Please wait...' : 'Create account'}
                </button>
              </form>
            )}
          </>
        )}
      </section>
    </main>
  )
}
