import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { useAuthStore } from '../../store/auth'
import { Button } from '../../components/ui/Button'
import { SignalRadar } from '../../components/signal/SignalRadar'
import { getGitHubAuthorizeUrl, getGitHubStatus, syncGitHub } from '../../api/github'
import { getSignalScores } from '../../api/signal'
import type { GitHubConnection, UserSignalScore } from '../../types'

// Static quest board — dimension-linked progression suggestions.
// These will become dynamic in a later phase.
const QUESTS = [
  {
    title: 'Connect GitHub',
    description: 'Link your GitHub account to seed your Builder, Executor, and Specialist signal.',
    dimensions: ['Builder', 'Executor', 'Specialist'] as const,
    cta: 'Connect →',
    action: 'github',
  },
  {
    title: 'Add a technical write-up',
    description: 'Publish analysis, a deep dive, or a how-to. Demonstrates your thinking, not just your output.',
    dimensions: ['Thinker'] as const,
    cta: 'Coming soon',
    action: null,
  },
  {
    title: 'Contribute to an OSS issue',
    description: 'A merged PR or meaningful issue response in a public repo adds Builder and Specialist evidence.',
    dimensions: ['Builder', 'Specialist'] as const,
    cta: 'Coming soon',
    action: null,
  },
  {
    title: 'Get peer feedback',
    description: 'A colleague or collaborator vouches for your work. Adds high-confidence Collaborator and Trusted signal.',
    dimensions: ['Collaborator', 'Trusted'] as const,
    cta: 'Coming soon',
    action: null,
  },
  {
    title: 'Log 3 weekly progress entries',
    description: 'Consistent tracked progress over time is the strongest Executor signal.',
    dimensions: ['Executor'] as const,
    cta: 'Coming soon',
    action: null,
  },
]

const DIM_COLOR: Record<string, string> = {
  Builder:      'text-amber-400  border-amber-400/30  bg-amber-400/5',
  Executor:     'text-sky-400    border-sky-400/30    bg-sky-400/5',
  Specialist:   'text-violet-400 border-violet-400/30 bg-violet-400/5',
  Thinker:      'text-emerald-400 border-emerald-400/30 bg-emerald-400/5',
  Collaborator: 'text-rose-400   border-rose-400/30   bg-rose-400/5',
  Trusted:      'text-gold-400   border-gold-400/30   bg-gold-400/5',
}

export function Hub() {
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()

  const [ghConn, setGhConn] = useState<GitHubConnection | null>(null)
  const [ghLoading, setGhLoading] = useState(true)
  const [ghConnecting, setGhConnecting] = useState(false)
  const [ghSyncing, setGhSyncing] = useState(false)
  const [ghError, setGhError] = useState<string | null>(null)
  const [ghJustConnected, setGhJustConnected] = useState(false)

  const [signal, setSignal] = useState<UserSignalScore | null>(null)

  function handleLogout() {
    logout()
    navigate('/login')
  }

  // Load GitHub status + signal scores on mount
  useEffect(() => {
    getGitHubStatus()
      .then(res => setGhConn(res.connection))
      .catch(() => setGhConn(null))
      .finally(() => setGhLoading(false))

    getSignalScores()
      .then(setSignal)
      .catch(() => {})
  }, [])

  // Handle redirect back from GitHub OAuth
  useEffect(() => {
    const status = searchParams.get('github')
    if (!status) return
    setSearchParams({}, { replace: true })

    if (status === 'connected') {
      setGhJustConnected(true)
      Promise.all([getGitHubStatus(), getSignalScores()])
        .then(([gh, sig]) => {
          setGhConn(gh.connection)
          setSignal(sig)
        })
        .catch(() => {})
    } else if (status === 'error') {
      setGhError('GitHub connection failed. Please try again.')
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  async function handleConnectGitHub() {
    setGhConnecting(true)
    setGhError(null)
    try {
      const url = await getGitHubAuthorizeUrl()
      window.location.href = url
    } catch {
      setGhError('Could not start GitHub connection.')
      setGhConnecting(false)
    }
  }

  async function handleSync() {
    setGhSyncing(true)
    setGhError(null)
    try {
      const [updated, scores] = await Promise.all([syncGitHub(), getSignalScores()])
      setGhConn(updated)
      setSignal(scores)
    } catch {
      setGhError('Sync failed. Please try again.')
    } finally {
      setGhSyncing(false)
    }
  }

  const hasBuild = false // dynamic in Phase 5

  const zeroSignal: UserSignalScore = {
    user_id: user?.id ?? 0,
    builder_score: 0, thinker_score: 0, executor_score: 0,
    collaborator_score: 0, specialist_score: 0, trusted_score: 0,
    total_signal: 0, updated_at: '',
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

      <main className="max-w-4xl mx-auto px-6 py-12 space-y-10">

        {/* Onboarding CTA */}
        {!hasBuild && (
          <div className="border border-gold-500/30 bg-void-900 p-8 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6">
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
        )}

        {/* Signal section */}
        <section className="border border-void-700 bg-void-900">
          <div className="border-b border-void-700 px-6 py-4 flex items-center justify-between">
            <span className="text-xs text-gold-400 tracking-[0.25em] uppercase">Signal</span>
            <span className="text-xs text-ink-600">multi-dimensional reputation</span>
          </div>
          <div className="p-6">
            <SignalRadar scores={signal ?? zeroSignal} />
          </div>
        </section>

        {/* GitHub evidence source */}
        <section className="border border-void-700 bg-void-900">
          <div className="border-b border-void-700 px-6 py-4 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <span className="text-xs text-gold-400 tracking-[0.25em] uppercase">Evidence Source</span>
              <span className="text-xs text-void-500">GitHub</span>
              {ghConn && (
                <span className="text-xs text-ink-500 font-mono">@{ghConn.github_username}</span>
              )}
            </div>
            {ghConn && (
              <button
                onClick={handleSync}
                disabled={ghSyncing}
                className="text-xs text-ink-500 hover:text-ink-300 transition-colors disabled:opacity-40"
              >
                {ghSyncing ? 'Syncing…' : '↻ Sync'}
              </button>
            )}
          </div>

          <div className="p-6">
            {ghError && <p className="text-red-400 text-xs mb-4">{ghError}</p>}
            {ghJustConnected && !ghError && (
              <p className="text-signal-400 text-xs mb-4 tracking-wide">◈ GitHub connected — signal updated</p>
            )}

            {ghLoading ? (
              <p className="text-ink-600 text-sm">Loading…</p>
            ) : ghConn ? (
              <GitHubStats conn={ghConn} />
            ) : (
              <GitHubConnectPrompt onConnect={handleConnectGitHub} connecting={ghConnecting} />
            )}
          </div>
        </section>

        {/* Quest board */}
        <section>
          <div className="flex items-center justify-between mb-4">
            <span className="text-xs text-gold-400 tracking-[0.25em] uppercase">Quest Board</span>
            <span className="text-xs text-ink-600">signal-improving actions</span>
          </div>
          <div className="grid grid-cols-1 gap-px bg-void-700 border border-void-700">
            {QUESTS.map(quest => {
              const githubConnected = !!ghConn
              const isGitHubQuest = quest.action === 'github'
              const done = isGitHubQuest && githubConnected
              return (
                <div
                  key={quest.title}
                  className={`bg-void-950 p-5 flex flex-col sm:flex-row sm:items-start gap-4 ${done ? 'opacity-50' : ''}`}
                >
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-sm font-medium text-ink-100">
                        {done ? '✓ ' : ''}{quest.title}
                      </span>
                      {quest.dimensions.map(dim => (
                        <span
                          key={dim}
                          className={`text-[10px] px-2 py-0.5 border rounded-none tracking-widest uppercase ${DIM_COLOR[dim]}`}
                        >
                          {dim}
                        </span>
                      ))}
                    </div>
                    <p className="text-xs text-ink-500 leading-relaxed">{quest.description}</p>
                  </div>
                  {!done && (
                    <div className="flex-shrink-0">
                      {isGitHubQuest ? (
                        <Button
                          onClick={handleConnectGitHub}
                          disabled={ghConnecting}
                          className="text-xs px-4 py-2 whitespace-nowrap"
                        >
                          {ghConnecting ? 'Redirecting…' : quest.cta}
                        </Button>
                      ) : (
                        <span className="text-xs text-ink-600">{quest.cta}</span>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </section>

      </main>
    </div>
  )
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function GitHubConnectPrompt({ onConnect, connecting }: { onConnect: () => void; connecting: boolean }) {
  return (
    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6">
      <div>
        <p className="text-sm text-ink-300 mb-1">Connect GitHub to add evidence</p>
        <p className="text-xs text-ink-600">
          Public repos, stars, top languages, and activity seed your Builder, Executor, and Specialist signal.
        </p>
      </div>
      <Button
        onClick={onConnect}
        disabled={connecting}
        className="flex-shrink-0 px-5 py-2 text-sm whitespace-nowrap"
      >
        {connecting ? 'Redirecting…' : 'Connect GitHub →'}
      </Button>
    </div>
  )
}

function GitHubStats({ conn }: { conn: GitHubConnection }) {
  const stats = [
    { label: 'Repos',     value: conn.repo_count },
    { label: 'Stars',     value: conn.star_count },
    { label: 'Followers', value: conn.followers },
  ]

  return (
    <div className="space-y-4">
      <div className="flex gap-8">
        {stats.map(s => (
          <div key={s.label} className="flex flex-col gap-0.5">
            <span className="text-xl font-bold text-ink-50 tabular-nums">{s.value}</span>
            <span className="text-[10px] text-ink-500 uppercase tracking-widest">{s.label}</span>
          </div>
        ))}
      </div>

      {conn.top_languages.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {conn.top_languages.map(lang => (
            <span
              key={lang}
              className="text-xs px-2.5 py-1 border border-void-600 text-ink-400 bg-void-800"
            >
              {lang}
            </span>
          ))}
        </div>
      )}

      {conn.synced_at && (
        <p className="text-xs text-ink-600">
          Last synced {new Date(conn.synced_at).toLocaleDateString('en-GB', {
            day: 'numeric', month: 'short', year: 'numeric',
          })}
        </p>
      )}
    </div>
  )
}
