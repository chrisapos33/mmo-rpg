import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getPublicProfile } from '../../api/profile'
import { SignalRadar } from '../../components/signal/SignalRadar'
import type { PublicProfileResponse, UserSignalScore } from '../../types'

const ZERO_SIGNAL: UserSignalScore = {
  user_id: 0,
  builder_score: 0, thinker_score: 0, executor_score: 0,
  collaborator_score: 0, specialist_score: 0, trusted_score: 0,
  total_signal: 0, updated_at: '',
}

export function PublicProfile() {
  const { userId } = useParams<{ userId: string }>()
  const [data, setData] = useState<PublicProfileResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    if (!userId) return
    getPublicProfile(userId)
      .then(setData)
      .catch(() => setNotFound(true))
      .finally(() => setLoading(false))
  }, [userId])

  if (loading) {
    return (
      <div className="min-h-screen bg-void-950 flex items-center justify-center">
        <span className="text-ink-600 text-sm tracking-widest animate-pulse">Loading…</span>
      </div>
    )
  }

  if (notFound || !data) {
    return (
      <div className="min-h-screen bg-void-950 flex flex-col items-center justify-center gap-6 px-6">
        <p className="text-gold-400 text-xs tracking-[0.4em] uppercase">Profile not found</p>
        <p className="text-ink-500 text-sm text-center max-w-xs">
          This profile doesn&rsquo;t exist or hasn&rsquo;t been published yet.
        </p>
        <Link to="/" className="text-xs text-ink-600 hover:text-ink-400 transition-colors">
          ← Return to home
        </Link>
      </div>
    )
  }

  const { profile, signal, github } = data
  const scores = signal ?? ZERO_SIGNAL

  return (
    <div className="min-h-screen bg-void-950 text-ink-50">

      {/* Minimal header */}
      <header className="border-b border-void-800 px-6 py-4 flex items-center justify-between">
        <Link to="/" className="text-gold-400 text-xs tracking-[0.35em] uppercase hover:text-gold-300 transition-colors">
          ◈ Signal
        </Link>
        <div className="flex items-center gap-4">
          <Link to="/explore" className="text-xs text-ink-500 hover:text-ink-300 transition-colors">
            Explore
          </Link>
          <Link
            to="/join"
            className="text-xs border border-void-600 px-4 py-2 text-ink-300 hover:border-gold-400 hover:text-gold-300 transition-colors"
          >
            Build yours →
          </Link>
        </div>
      </header>

      {/* Hero */}
      <section className="border-b border-void-800 px-6 py-16 text-center relative overflow-hidden">
        {/* Background glow */}
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            background: 'radial-gradient(ellipse 60% 40% at 50% 60%, rgba(180,140,60,0.06) 0%, transparent 70%)',
          }}
        />

        <div className="relative max-w-2xl mx-auto">
          <p className="text-gold-400 text-xs tracking-[0.5em] uppercase mb-6">
            ◈ &nbsp; Signal Profile
          </p>
          <h1 className="text-5xl sm:text-6xl md:text-7xl font-black tracking-tight text-gold-400 leading-none mb-3">
            {profile.class}
          </h1>
          <p className="text-sm text-ink-400 tracking-[0.25em] uppercase mb-6">
            {profile.subclass}
          </p>
          {profile.headline && (
            <p className="text-base text-ink-300 italic max-w-xl mx-auto leading-relaxed">
              &ldquo;{profile.headline}&rdquo;
            </p>
          )}
          {github && (
            <div className="mt-6 inline-flex items-center gap-2 text-xs text-ink-500 border border-void-700 px-3 py-1.5">
              <span className="text-ink-600">github.com/</span>
              <span className="text-ink-300 font-mono">{github.github_username}</span>
              <span className="text-void-600">·</span>
              <span className="text-signal-400 text-[10px] tracking-widest uppercase">Platform Verified</span>
            </div>
          )}
        </div>
      </section>

      {/* Main content */}
      <main className="max-w-4xl mx-auto px-6 py-12 space-y-12">

        {/* Signal radar + summary/strengths */}
        <section className="grid lg:grid-cols-2 gap-12 items-start">
          {/* Left: radar */}
          <div>
            <p className="text-xs text-gold-400 tracking-[0.25em] uppercase mb-6">Signal</p>
            <SignalRadar scores={scores} />
          </div>

          {/* Right: summary + traits */}
          <div className="space-y-8">
            {profile.summary && (
              <div>
                <p className="text-xs text-gold-400 tracking-[0.25em] uppercase mb-3">About</p>
                <p className="text-sm text-ink-400 leading-relaxed">{profile.summary}</p>
              </div>
            )}

            {(profile.strengths ?? []).length > 0 && (
              <div>
                <p className="text-xs text-gold-400 tracking-[0.25em] uppercase mb-3">Strengths</p>
                <ul className="space-y-2">
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
                <p className="text-xs text-gold-400 tracking-[0.25em] uppercase mb-3">Growth Path</p>
                <ul className="space-y-2">
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
        </section>

        {/* GitHub evidence block */}
        {github && (
          <section className="border border-void-700 bg-void-900 p-6">
            <div className="flex items-center justify-between mb-5">
              <div className="flex items-center gap-3">
                <span className="text-xs text-gold-400 tracking-[0.25em] uppercase">GitHub Evidence</span>
                <span className="text-[10px] text-signal-400 border border-signal-400/30 px-2 py-0.5 tracking-widest uppercase">
                  Platform Verified
                </span>
              </div>
              <a
                href={`https://github.com/${github.github_username}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-ink-500 hover:text-ink-300 transition-colors font-mono"
              >
                @{github.github_username} ↗
              </a>
            </div>

            <div className="flex gap-10 mb-5">
              {[
                { label: 'Repos',     value: github.repo_count },
                { label: 'Stars',     value: github.star_count },
                { label: 'Followers', value: github.followers },
              ].map(s => (
                <div key={s.label} className="flex flex-col gap-0.5">
                  <span className="text-2xl font-bold text-ink-50 tabular-nums">{s.value}</span>
                  <span className="text-[10px] text-ink-500 uppercase tracking-widest">{s.label}</span>
                </div>
              ))}
            </div>

            {github.top_languages.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {github.top_languages.map(lang => (
                  <span
                    key={lang}
                    className="text-xs px-2.5 py-1 border border-void-600 text-ink-400 bg-void-800"
                  >
                    {lang}
                  </span>
                ))}
              </div>
            )}
          </section>
        )}

      </main>

      {/* Footer CTA */}
      <footer className="border-t border-void-800 px-6 py-12 text-center">
        <p className="text-xs text-ink-600 tracking-widest uppercase mb-3">
          Multi-dimensional developer signal
        </p>
        <p className="text-sm text-ink-400 mb-6 max-w-sm mx-auto">
          Build your own verified signal profile. Connect your evidence, claim your class.
        </p>
        <Link
          to="/join"
          className="inline-flex items-center gap-2 bg-gold-400 text-void-950 text-sm font-medium px-8 py-3 hover:bg-gold-300 transition-colors"
        >
          Build your profile →
        </Link>
        <p className="mt-6 text-xs text-void-600">
          Powered by Signal
        </p>
      </footer>

    </div>
  )
}
