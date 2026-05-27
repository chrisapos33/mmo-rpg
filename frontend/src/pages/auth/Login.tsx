import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { authApi } from '../../api/auth'
import { ApiError } from '../../api/client'
import { useAuthStore } from '../../store/auth'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'

export function Login() {
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
      const { token, user } = await authApi.login(email, password)
      setAuth(token, user)
      navigate('/hub')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-void-950 flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <div className="mb-10 text-center">
          <div className="text-gold-400 text-xs tracking-[0.3em] uppercase mb-2">Welcome Back</div>
          <h1 className="text-2xl font-semibold text-ink-50 tracking-tight">Sign in</h1>
          <p className="mt-2 text-sm text-ink-400">Continue your journey.</p>
        </div>

        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <Input
            label="Email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={e => setEmail(e.target.value)}
            autoComplete="email"
            required
          />
          <Input
            label="Password"
            type="password"
            placeholder="Your password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            autoComplete="current-password"
            required
          />

          {error && (
            <p className="text-sm text-danger-400 text-center">{error}</p>
          )}

          <Button type="submit" loading={loading} className="w-full mt-1">
            Enter
          </Button>
        </form>

        <p className="mt-8 text-center text-sm text-ink-400">
          No identity yet?{' '}
          <Link to="/join" className="text-gold-400 hover:text-gold-300 transition-colors">
            Join the Hunt
          </Link>
        </p>
      </div>
    </div>
  )
}
