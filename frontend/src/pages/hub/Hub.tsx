import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { useAuthStore } from '../../store/auth'
import { Button } from '../../components/ui/Button'
import { SignalRadar } from '../../components/signal/SignalRadar'
import { getGitHubAuthorizeUrl, getGitHubStatus, syncGitHub } from '../../api/github'
import { getSignalScores } from '../../api/signal'
import { getBuild } from '../../api/onboarding'
import { publishProfile } from '../../api/profile'
import { ApiError } from '../../api/client'
import type { GitHubConnection, UserSignalScore, Profile } from '../../types'

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
  const [build, setBuild] = useState<Profile | null>(null)
  const [publishing, setPublishing] = useState(false)
  const [publishError, setPublishError] = useState<string | null>(null)

  function handleLogout() {
    logout()
    navigate('/login')
  }

  // Load build, GitHub status, and signal scores on mount
  useEffect(() => {
    getBuild()
      .then(setBuild)
      .catch(err => {
        if (!(err instanceof ApiError && err.status === 404)) console.error(err)
      })

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

  const hasBuild = build !== null

  async function handlePublish() {
    setPublishing(true)
    setPublishError(null)
    try {
      await publishProfile()
      // Refetch build to get updated is_published
      const updated = await getBuild()
      setBuild(updated)
    } catch {
      setPublishError('Publish failed. Please try again.')
    } finally {
      setPublishing(false)
    }
  }

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

        {/* Character build panel */}
        {build && (
          <BuildPanel
            profile={build}
            onPublish={handlePublish}
            publishing={publishing}
            publishError={publishError}
          />
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

function BuildPanel({
  profile,
  onPublish,
  publishing,
  publishError,
}: {
  profile: Profile
  onPublish: () => void
  publishing: boolean
  publishError: string | null
}) {
  const profileUrl = `/p/${profile.user_id}`

  function copyUrl() {
    navigator.clipboard.writeText(window.location.origin + profileUrl)
  }

  return (
    <section className="border border-void-700 bg-void-900">
      <div className="border-b border-void-700 px-6 py-4 flex items-center justify-between">
        <span className="text-xs text-gold-400 tracking-[0.25em] uppercase">Character Build</span>
        <Link
          to="/onboarding/reveal"
          className="text-xs text-ink-600 hover:text-ink-400 transition-colors"
        >
          Rebuild →
        </Link>
      </div>

      <div className="p-6">
        {/* Class + subclass */}
        <div className="mb-6">
          <h2 className="text-4xl font-black tracking-tight text-gold-400 leading-none">
            {profile.class}
          </h2>
          <p className="mt-1.5 text-sm text-ink-400 tracking-[0.2em] uppercase">
            {profile.subclass}
          </p>
        </div>

        {/* Headline */}
        {profile.headline && (
          <p className="text-sm text-ink-300 italic border-l-2 border-gold-400/30 pl-4 mb-6 leading-relaxed">
            &ldquo;{profile.headline}&rdquo;
          </p>
        )}

        {/* Summary */}
        {profile.summary && (
          <p className="text-sm text-ink-500 leading-relaxed mb-6">
            {profile.summary}
          </p>
        )}

        {/* Strengths + growth paths */}
        <div className="grid sm:grid-cols-2 gap-6 mb-8">
          {(profile.strengths ?? []).length > 0 && (
            <div>
              <p className="text-[10px] text-gold-400 tracking-[0.3em] uppercase mb-3">Strengths</p>
              <ul className="space-y-1.5">
                {profile.strengths.map((s, i) => (
                  <li key={i} className="flex items-start gap-2 text-sm text-ink-300">
                    <span className="text-gold-400 mt-0.5 flex-shrink-0 text-xs">◆</span>
                    {s}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {(profile.growth_paths ?? []).length > 0 && (
            <div>
              <p className="text-[10px] text-gold-400 tracking-[0.3em] uppercase mb-3">Growth Path</p>
              <ul className="space-y-1.5">
                {profile.growth_paths.map((g, i) => (
                  <li key={i} className="flex items-start gap-2 text-sm text-ink-400">
                    <span className="text-ink-600 mt-0.5 flex-shrink-0">→</span>
                    {g}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Publish / share */}
        <div className="border-t border-void-700 pt-6">
          {publishError && (
            <p className="text-red-400 text-xs mb-3">{publishError}</p>
          )}
          {profile.is_published ? (
            <div className="flex flex-col sm:flex-row sm:items-center gap-4">
              <div className="flex-1">
                <p className="text-[10px] text-signal-400 tracking-widest uppercase mb-1">Profile published</p>
                <p className="text-xs text-ink-500 font-mono">
                  {window.location.origin}{profileUrl}
                </p>
              </div>
              <div className="flex items-center gap-3 flex-shrink-0">
                <button
                  onClick={copyUrl}
                  className="text-xs border border-void-600 px-4 py-2 text-ink-300 hover:border-gold-400 hover:text-gold-300 transition-colors"
                >
                  Copy link
                </button>
                <Link
                  to={profileUrl}
                  target="_blank"
                  className="text-xs text-ink-600 hover:text-ink-400 transition-colors"
                >
                  View ↗
                </Link>
              </div>
            </div>
          ) : (
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <p className="text-sm text-ink-300 mb-0.5">Share your signal profile</p>
                <p className="text-xs text-ink-600">
                  Make your character build and signal publicly accessible via a link.
                </p>
              </div>
              <Button
                onClick={onPublish}
                disabled={publishing}
                loading={publishing}
                className="flex-shrink-0 px-6 py-2 text-sm whitespace-nowrap"
              >
                Publish profile
              </Button>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

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
