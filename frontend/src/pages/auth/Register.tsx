import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authApi } from '../../api/auth'
import { ApiError } from '../../api/client'
import { useAuthStore } from '../../store/auth'

const FANTASY_FONT = { fontFamily: '"Cinzel", serif' }

export function Register() {
  const navigate = useNavigate()
  const setAuth = useAuthStore(s => s.setAuth)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const { token, user } = await authApi.register(email, password)
      setAuth(token, user)
      navigate('/hub')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-void-950 flex flex-col items-center justify-center px-4 relative overflow-hidden">
      {/* Ambient glow */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          background: 'radial-gradient(ellipse 60% 50% at 50% 50%, rgba(160,112,24,0.08) 0%, transparent 70%)',
        }}
      />

      {/* Logo */}
      <Link
        to="/"
        className="mb-10 text-gold-400 text-xs tracking-[0.4em] uppercase hover:text-gold-300 transition-colors"
        style={FANTASY_FONT}
      >
        ◈ Signal
      </Link>

      {/* Panel */}
      <div
        className="relative w-full max-w-sm p-8"
        style={{ border: '1px solid rgba(74,64,50,0.5)' }}
      >
        {/* Corner diamonds */}
        {[
          { top: -4, left: -4 },
          { top: -4, right: -4 },
          { bottom: -4, left: -4 },
          { bottom: -4, right: -4 },
        ].map((pos, i) => (
          <svg
            key={i}
            width="8" height="8" viewBox="0 0 10 10"
            className="absolute"
            style={{ ...pos }}
            aria-hidden="true"
          >
            <path d="M5 0L10 5L5 10L0 5Z" fill="rgba(200,150,50,0.9)" />
          </svg>
        ))}

        <div className="mb-8 text-center">
          <p
            className="text-[10px] tracking-[0.5em] uppercase mb-3"
            style={{ ...FANTASY_FONT, color: 'rgba(160,112,24,0.65)' }}
          >
            The Hunt Begins
          </p>
          <h1
            className="text-2xl tracking-wide text-ink-50"
            style={FANTASY_FONT}
          >
            Forge Your Identity
          </h1>
          <p className="mt-2 text-xs text-ink-400">Your professional build starts here.</p>
        </div>

        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <div className="flex flex-col gap-1.5">
            <label
              className="text-[10px] tracking-[0.35em] uppercase"
              style={{ color: 'rgba(160,112,24,0.7)' }}
            >
              Email
            </label>
            <input
              type="email"
              placeholder="you@example.com"
              value={email}
              onChange={e => setEmail(e.target.value)}
              autoComplete="email"
              required
              className="w-full px-4 py-2.5 text-sm text-ink-50 placeholder:text-ink-600 outline-none transition-colors"
              style={{
                background: 'rgba(26,18,8,0.8)',
                border: '1px solid rgba(74,64,50,0.5)',
              }}
              onFocus={e => (e.target.style.borderColor = 'rgba(160,112,24,0.6)')}
              onBlur={e => (e.target.style.borderColor = 'rgba(74,64,50,0.5)')}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label
              className="text-[10px] tracking-[0.35em] uppercase"
              style={{ color: 'rgba(160,112,24,0.7)' }}
            >
              Password
            </label>
            <input
              type="password"
              placeholder="Minimum 8 characters"
              value={password}
              onChange={e => setPassword(e.target.value)}
              autoComplete="new-password"
              required
              className="w-full px-4 py-2.5 text-sm text-ink-50 placeholder:text-ink-600 outline-none transition-colors"
              style={{
                background: 'rgba(26,18,8,0.8)',
                border: '1px solid rgba(74,64,50,0.5)',
              }}
              onFocus={e => (e.target.style.borderColor = 'rgba(160,112,24,0.6)')}
              onBlur={e => (e.target.style.borderColor = 'rgba(74,64,50,0.5)')}
            />
          </div>

          {error && (
            <p className="text-sm text-danger-400 text-center">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full mt-1 py-3 text-sm font-medium tracking-[0.08em] transition-opacity hover:opacity-90 disabled:opacity-50"
            style={{
              ...FANTASY_FONT,
              background: 'linear-gradient(135deg, #c99224 0%, #8c6018 100%)',
              color: '#130e07',
            }}
          >
            {loading ? (
              <span className="inline-flex items-center gap-2 justify-center">
                <span className="size-3.5 rounded-full border-2 border-current border-t-transparent animate-spin" />
                Creating…
              </span>
            ) : (
              'Join the Hunt →'
            )}
          </button>
        </form>

        <div className="mt-6 pt-6 text-center" style={{ borderTop: '1px solid rgba(74,64,50,0.3)' }}>
          <p className="text-xs text-ink-400">
            Already have an identity?{' '}
            <Link to="/login" className="text-gold-400 hover:text-gold-300 transition-colors">
              Sign in
            </Link>
          </p>
        </div>
      </div>

      {/* Trust footnote */}
      <p
        className="mt-8 text-[10px] tracking-widest uppercase"
        style={{ color: 'rgba(74,64,50,0.6)' }}
      >
        Scored from real GitHub activity · Not self-reported
      </p>
    </div>
  )
}
