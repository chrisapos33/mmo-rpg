import { useAuthStore } from '../../store/auth'
import { Button } from '../../components/ui/Button'
import { useNavigate } from 'react-router-dom'

export function Hub() {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-void-950 text-ink-50">
      {/* Top bar */}
      <header className="border-b border-void-700 px-6 py-4 flex items-center justify-between">
        <div>
          <span className="text-gold-400 text-xs tracking-[0.3em] uppercase">Hunter HQ</span>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-ink-400">{user?.email}</span>
          <Button variant="ghost" onClick={handleLogout} className="text-xs px-4 py-2">
            Sign out
          </Button>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-4xl mx-auto px-6 py-16">
        <div className="text-center mb-12">
          <p className="text-gold-400 text-xs tracking-[0.3em] uppercase mb-3">Phase 0 — Foundation</p>
          <h2 className="text-3xl font-semibold text-ink-50 tracking-tight">Identity established.</h2>
          <p className="mt-3 text-ink-400 max-w-md mx-auto">
            Authentication is live. Your professional build will take shape here as we complete the onboarding flow.
          </p>
        </div>

        {/* Placeholder panels */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[
            { label: 'Character Build', status: 'Coming in Phase 2' },
            { label: 'Signal Score',    status: 'Coming in Phase 4' },
            { label: 'Quest Board',     status: 'Coming in Phase 5' },
          ].map(panel => (
            <div
              key={panel.label}
              className="border border-void-700 bg-void-900 p-6 flex flex-col gap-2"
            >
              <span className="text-xs text-gold-400 tracking-widest uppercase">{panel.label}</span>
              <span className="text-sm text-ink-600">{panel.status}</span>
            </div>
          ))}
        </div>
      </main>
    </div>
  )
}
