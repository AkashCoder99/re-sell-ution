import { useEffect, useMemo, useState } from 'react'
import type { ChangeEvent, FormEvent } from 'react'
import { getMe, login, logout, register, updateProfile } from './api/auth'
import type { LoginRequest, RegisterRequest } from './api/auth'
import type { PublicUser, UpdateProfileRequest } from './types/user'
import CitySelector from './components/CitySelector'
import ProfileEdit from './components/ProfileEdit'
import ForgotPassword from './components/ForgotPassword'
import ResetPassword from './components/ResetPassword'
import CreateListing from './components/CreateListing'
import MyListingsDashboard from './components/MyListingsDashboard'
import {
  IconEmail,
  IconLocation,
  IconBio,
  IconImage,
  IconAddListing,
  IconListings,
  IconEdit,
  IconCity,
  IconLogout,
  IconBack
} from './components/Icons'

const defaultLoginForm: LoginRequest = { email: '', password: '' }
const defaultRegisterForm: RegisterRequest = { full_name: '', email: '', password: '' }

/** Preview mode: use this token + user to explore the app without logging in (mock API only). */
const PREVIEW_TOKEN = 'preview_no_login'
const PREVIEW_USER: PublicUser = {
  id: 'preview_user',
  email: 'preview@resellution.demo',
  full_name: 'Demo User',
  city: 'New York'
}

type ViewMode =
  | 'login'
  | 'register'
  | 'forgot-password'
  | 'reset-password'
  | 'city-select'
  | 'profile'
  | 'profile-edit'
  | 'create-listing'
  | 'my-listings'

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
  const [listingsRefresh, setListingsRefresh] = useState(0)
  const isAuthenticated = useMemo(() => Boolean(token && user), [token, user])

  useEffect(() => {
    if (!token) {
      setUser(null)
      return
    }

    if (token === PREVIEW_TOKEN) {
      setUser(PREVIEW_USER)
      setShowCitySelector(false)
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
    if (token === PREVIEW_TOKEN) {
      setToken('')
      setUser(null)
      setViewMode('login')
      setShowCitySelector(false)
      setMessage('Preview ended.')
      return
    }

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

    if (token === PREVIEW_TOKEN) {
      setUser((prev) => (prev ? { ...prev, city } : null))
      setShowCitySelector(false)
      setViewMode('profile')
      setMessage(`City set to ${city} (preview)`)
      return
    }

    const data = await updateProfile(token, { city })
    setUser(data.user)
    setShowCitySelector(false)
    setViewMode('profile')
    setMessage(`City updated to ${city}`)
  }

  async function handleProfileUpdate(updates: UpdateProfileRequest) {
    if (!token) return

    if (token === PREVIEW_TOKEN) {
      setUser((prev) => (prev ? { ...prev, ...updates } : null))
      setViewMode('profile')
      setMessage('Profile updated (preview)')
      return
    }

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
            {token === PREVIEW_TOKEN && (
              <p className="preview-banner">You are in preview mode — no real account. Data is mock-only.</p>
            )}
            <h2 className="profile-welcome">Welcome, {user.full_name}</h2>
            <div className="profile-details">
              <div className="profile-detail-row">
                <IconEmail className="profile-detail-icon" aria-hidden />
                <span className="profile-detail-label">Email</span>
                <span className="profile-detail-value">{user.email}</span>
              </div>
              {user.city && (
                <div className="profile-detail-row">
                  <IconLocation className="profile-detail-icon" aria-hidden />
                  <span className="profile-detail-label">City</span>
                  <span className="profile-detail-value">{user.city}</span>
                </div>
              )}
              {user.bio && (
                <div className="profile-detail-row">
                  <IconBio className="profile-detail-icon" aria-hidden />
                  <span className="profile-detail-label">Bio</span>
                  <span className="profile-detail-value">{user.bio}</span>
                </div>
              )}
              {user.photo_url && (
                <div className="profile-detail-row">
                  <IconImage className="profile-detail-icon" aria-hidden />
                  <span className="profile-detail-label">Photo</span>
                  <span className="profile-detail-value">
                    <a href={user.photo_url} target="_blank" rel="noopener noreferrer">View</a>
                  </span>
                </div>
              )}
            </div>
            <div className="profile-actions">
              <button type="button" className="profile-action-btn primary" onClick={() => setViewMode('create-listing')}>
                <IconAddListing className="profile-action-icon" aria-hidden />
                <span>Create Listing</span>
              </button>
              <button type="button" className="profile-action-btn primary" onClick={() => setViewMode('my-listings')}>
                <IconListings className="profile-action-icon" aria-hidden />
                <span>My Listings</span>
              </button>
              <button type="button" className="profile-action-btn" onClick={() => setViewMode('profile-edit')}>
                <IconEdit className="profile-action-icon" aria-hidden />
                <span>Edit Profile</span>
              </button>
              <button type="button" className="profile-action-btn" onClick={() => setShowCitySelector(true)}>
                <IconCity className="profile-action-icon" aria-hidden />
                <span>Change City</span>
              </button>
              <button type="button" className="profile-action-btn secondary" onClick={handleLogout}>
                <IconLogout className="profile-action-icon" aria-hidden />
                <span>Logout</span>
              </button>
            </div>
          </div>
        ) : isAuthenticated && viewMode === 'create-listing' && user ? (
          <CreateListing
            token={token}
            userCity={user.city || ''}
            onSuccess={() => {
              setViewMode('profile')
              setListingsRefresh((r) => r + 1)
              setMessage('Listing created successfully.')
            }}
            onCancel={() => setViewMode('profile')}
          />
        ) : isAuthenticated && viewMode === 'my-listings' ? (
          <>
            <button
              type="button"
              className="back-link-btn"
              onClick={() => setViewMode('profile')}
            >
              <IconBack className="back-link-icon" aria-hidden />
              <span>Back to Profile</span>
            </button>
            <MyListingsDashboard
              token={token}
              refreshTrigger={listingsRefresh}
              onMarkSold={() => setListingsRefresh((r) => r + 1)}
            />
          </>
        ) : viewMode === 'forgot-password' ? (
          /* Forgot Password */
          <ForgotPassword onBack={() => setViewMode('login')} onGoToReset={() => setViewMode('reset-password')} />
        ) : viewMode === 'reset-password' ? (
          /* Reset Password */
          <ResetPassword onBack={() => setViewMode('login')} />
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

                <button
                  type="button"
                  className="link-button preview-button"
                  onClick={() => {
                    setToken(PREVIEW_TOKEN)
                    setUser(PREVIEW_USER)
                    setViewMode('profile')
                    setMessage('Preview mode — explore without an account. Use mock data.')
                  }}
                >
                  Preview app without login
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

                <button
                  type="button"
                  className="link-button preview-button"
                  onClick={() => {
                    setToken(PREVIEW_TOKEN)
                    setUser(PREVIEW_USER)
                    setViewMode('profile')
                    setMessage('Preview mode — explore without an account. Use mock data.')
                  }}
                >
                  Preview app without login
                </button>
              </form>
            )}
          </>
        )}
      </section>
    </main>
  )
}
