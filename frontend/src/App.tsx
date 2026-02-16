import { useEffect, useMemo, useState } from 'react'
import type { ChangeEvent, FormEvent } from 'react'
import { getMe, login, logout, register, updateProfile } from './api/auth'
import type { LoginRequest, RegisterRequest } from './api/auth'
import type { PublicUser, UpdateProfileRequest } from './types/user'
import CitySelector from './components/CitySelector'
import ProfileEdit from './components/ProfileEdit'
import ForgotPassword from './components/ForgotPassword'

const defaultLoginForm: LoginRequest = { email: '', password: '' }
const defaultRegisterForm: RegisterRequest = { full_name: '', email: '', password: '' }

type ViewMode = 'login' | 'register' | 'forgot-password' | 'city-select' | 'profile' | 'profile-edit'

export default function App() {
  const [viewMode, setViewMode] = useState<ViewMode>('login')
  const [loginForm, setLoginForm] = useState<LoginRequest>(defaultLoginForm)
  const [registerForm, setRegisterForm] = useState<RegisterRequest>(defaultRegisterForm)
  const [token, setToken] = useState<string>(() => localStorage.getItem('auth_token') || '')
  const [user, setUser] = useState<PublicUser | null>(null)
  const [loading, setLoading] = useState<boolean>(false)
  const [message, setMessage] = useState<string>('')
  const [showCitySelector, setShowCitySelector] = useState<boolean>(false)
  const [showLoginPassword, setShowLoginPassword] = useState<boolean>(false)
  const [showRegisterPassword, setShowRegisterPassword] = useState<boolean>(false)
  const isAuthenticated = useMemo(() => Boolean(token && user), [token, user])

  useEffect(() => {
    if (!token) {
      setUser(null)
      return
    }

    getMe(token)
      .then((data) => {
        setUser(data.user)
        // If user doesn't have a city set, prompt them to select one
        if (!data.user.city) {
          setShowCitySelector(true)
        }
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
      
      // Check if user needs to select city
      if (!data.user.city) {
        setShowCitySelector(true)
      } else {
        setViewMode('profile')
      }
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
      
      // Prompt new user to select city
      setShowCitySelector(true)
    } catch (error: unknown) {
      setMessage(error instanceof Error ? error.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  async function handleLogout() {
    setLoading(true)
    setMessage('')

    try {
      if (token) {
        await logout(token)
      }
      setMessage('Logged out.')
    } catch (error: unknown) {
      setMessage(
        error instanceof Error
          ? `Logout failed on server; signed out locally: ${error.message}`
          : 'Logout failed on server; signed out locally.'
      )
    } finally {
      localStorage.removeItem('auth_token')
      setToken('')
      setUser(null)
      setViewMode('login')
      setShowCitySelector(false)
      setLoading(false)
    }
  }

  async function handleCitySelected(city: string) {
    if (!token) return

    const data = await updateProfile(token, { city })
    setUser(data.user)
    setShowCitySelector(false)
    setViewMode('profile')
    setMessage(`City updated to ${city}`)
  }

  async function handleProfileUpdate(updates: UpdateProfileRequest) {
    if (!token) return

    const data = await updateProfile(token, updates)
    setUser(data.user)
    setViewMode('profile')
    setMessage('Profile updated successfully')
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
        <h1>🛍️ ReSellution</h1>
        <p className="subtitle">Your local marketplace platform</p>

        {message && <p className="message">{message}</p>}

        {/* City Selector - shown after signup/login if no city set */}
        {isAuthenticated && showCitySelector ? (
          <CitySelector
            currentCity={user?.city}
            onCitySelected={handleCitySelected}
            onSkip={() => {
              setShowCitySelector(false)
              setViewMode('profile')
            }}
          />
        ) : isAuthenticated && viewMode === 'profile-edit' && user ? (
          /* Profile Edit Mode */
          <ProfileEdit
            user={user}
            onUpdate={handleProfileUpdate}
            onCancel={() => setViewMode('profile')}
          />
        ) : isAuthenticated && user ? (
          /* Profile View */
          <div className="profile">
            <h2>👋 Welcome, {user.full_name}</h2>
            <div className="profile-details">
              <p><strong>📧 Email:</strong> {user.email}</p>
              {user.city && <p><strong>📍 City:</strong> {user.city}</p>}
              {user.bio && <p><strong>💬 Bio:</strong> {user.bio}</p>}
              {user.photo_url && <p><strong>🖼️ Photo:</strong> <a href={user.photo_url} target="_blank" rel="noopener noreferrer">View</a></p>}
            </div>
            <div className="button-group">
              <button type="button" onClick={() => setViewMode('profile-edit')}>
                Edit Profile
              </button>
              <button type="button" onClick={() => setShowCitySelector(true)}>
                Change City
              </button>
              <button type="button" className="secondary" onClick={handleLogout}>
                Logout
              </button>
            </div>
          </div>
        ) : viewMode === 'forgot-password' ? (
          /* Forgot Password */
          <ForgotPassword onBack={() => setViewMode('login')} />
        ) : (
          /* Login/Register */
          <>
            <div className="mode-toggle">
              <button
                type="button"
                className={viewMode === 'login' ? 'active' : ''}
                onClick={() => setViewMode('login')}
              >
                Login
              </button>
              <button
                type="button"
                className={viewMode === 'register' ? 'active' : ''}
                onClick={() => setViewMode('register')}
              >
                Register
              </button>
            </div>

            {viewMode === 'login' ? (
              <form onSubmit={handleLogin} className="auth-form">
                <label>
                  Email
                  <input type="email" value={loginForm.email} onChange={onLoginEmailChange} required />
                </label>

                <label>
                  Password
                  <div className="password-input-wrapper">
                    <input 
                      type={showLoginPassword ? "text" : "password"} 
                      value={loginForm.password} 
                      onChange={onLoginPasswordChange} 
                      required 
                    />
                    <button
                      type="button"
                      className="password-toggle"
                      onClick={() => setShowLoginPassword(!showLoginPassword)}
                      aria-label={showLoginPassword ? "Hide password" : "Show password"}
                    >
                      {showLoginPassword ? '👁️' : '👁️‍🗨️'}
                    </button>
                  </div>
                </label>

                <button type="submit" disabled={loading}>
                  {loading ? 'Please wait...' : 'Login'}
                </button>

                <button
                  type="button"
                  className="link-button"
                  onClick={() => setViewMode('forgot-password')}
                >
                  Forgot password?
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
                  <div className="password-input-wrapper">
                    <input
                      type={showRegisterPassword ? "text" : "password"}
                      value={registerForm.password}
                      onChange={onRegisterPasswordChange}
                      required
                      minLength={8}
                    />
                    <button
                      type="button"
                      className="password-toggle"
                      onClick={() => setShowRegisterPassword(!showRegisterPassword)}
                      aria-label={showRegisterPassword ? "Hide password" : "Show password"}
                    >
                      {showRegisterPassword ? '👁️' : '👁️‍🗨️'}
                    </button>
                  </div>
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
