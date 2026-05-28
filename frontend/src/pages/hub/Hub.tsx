import { useNavigate, Link } from 'react-router-dom'
import { useAuthStore } from '../../store/auth'
import { Button } from '../../components/ui/Button'

export function Hub() {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-void-950 text-ink-50">
      <header className="border-b border-void-700 px-6 py-4 flex items-center justify-between">
        <span className="text-gold-400 text-xs tracking-[0.3em] uppercase">
          <span className="text-ink-50">◈</span> Hunter HQ
        </span>
        <div className="flex items-center gap-4">
          <span className="text-sm text-ink-400">{user?.email}</span>
          <Button variant="ghost" onClick={handleLogout} className="text-xs px-4 py-2">
            Sign out
          </Button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-16">
        {/* Onboarding prompt */}
        <div className="border border-gold-500/30 bg-void-900 p-8 mb-10 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6">
          <div>
            <p className="text-gold-400 text-xs tracking-[0.3em] uppercase mb-2">Your build is incomplete</p>
            <h2 className="text-lg font-semibold text-ink-50">Upload your CV to begin</h2>
            <p className="mt-1 text-sm text-ink-400">
              AI will analyze your experience and initialize your professional identity.
            </p>
          </div>
          <Link to="/onboarding/upload" className="flex-shrink-0">
            <Button className="px-6 py-2.5 text-sm whitespace-nowrap">Start onboarding →</Button>
          </Link>
        </div>

        {/* Panel grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-px bg-void-700 border border-void-700">
          {[
            { label: 'Character Build', detail: 'Upload CV to initialize', phase: '02' },
            { label: 'Signal Score',    detail: 'Connect GitHub to begin',  phase: '04' },
            { label: 'Quest Board',     detail: 'Complete onboarding first', phase: '05' },
          ].map(panel => (
            <div key={panel.label} className="bg-void-950 p-6 flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <span className="text-xs text-gold-400 tracking-widest uppercase">{panel.label}</span>
                <span className="text-xs text-void-600 font-mono">Phase {panel.phase}</span>
              </div>
              <span className="text-sm text-ink-600">{panel.detail}</span>
            </div>
          ))}
        </div>
      </main>
    </div>
  )
}
