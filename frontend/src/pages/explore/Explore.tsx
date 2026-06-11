import { useEffect, useState, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { getExplore } from '../../api/explore'
import { useAuthStore } from '../../store/auth'
import type { ExploreEntry } from '../../types'

const FANTASY_FONT = { fontFamily: '"Cinzel", serif' }

const ALL_CLASSES = [
  'The Architect',
  'The Artisan',
  'The Pathfinder',
  'The Sage',
  'The Operator',
  'The Sentinel',
  'The Artificer',
]

type Sort = 'signal' | 'recent'

// Small inline class icon for cards and pills
function ClassIcon({ cls, size = 14 }: { cls: string; size?: number }) {
  const paths: Record<string, string> = {
    'The Architect':  'M6 2l4 3v7H2V5l4-3zm0 2L3 6v5h6V6L6 4zm0 3a1 1 0 110 2 1 1 0 010-2z',
    'The Artisan':    'M3 9l3-6 3 6H3zm3-4.5L4 9h4L6 4.5zM9 9a3 3 0 100 6 3 3 0 000-6zm0 1a2 2 0 110 4 2 2 0 010-4z',
    'The Pathfinder': 'M6 1l1.5 4H11l-2.8 2 1 3.5L6 8.5 2.8 10.5l1-3.5L1 5h3.5L6 1z',
    'The Sage':       'M6 1a5 5 0 100 10A5 5 0 006 1zm0 1a4 4 0 110 8A4 4 0 016 2zm-1 2v5h2V6h1V4H5z',
    'The Operator':   'M2 3h8v1H2V3zm0 2h8v1H2V5zm0 2h5v1H2V7zm7 0l2 2-2 2v-4z',
    'The Sentinel':   'M6 1L2 3v4c0 2.5 1.8 4.6 4 5 2.2-.4 4-2.5 4-5V3L6 1zm0 1.5l3 1.5V7c0 1.8-1.3 3.3-3 3.8C4.3 10.3 3 8.8 3 7V4l3-1.5z',
    'The Artificer':  'M4 2l2 2-2 2-2-2 2-2zm4 0l2 2-2 2-2-2 2-2zm-2 4l2 2-2 2-2-2 2-2zm-4 4l2 2-2 2-2-2 2-2zm4 0l2 2-2 2-2-2 2-2zm4 0l2 2-2 2-2-2 2-2z',
  }
  const d = paths[cls] ?? paths['The Architect']
  return (
    <svg width={size} height={size} viewBox="0 0 12 12" fill="none" aria-hidden="true">
      <path d={d} fill="currentColor" fillRule="evenodd" />
    </svg>
  )
}

export function Explore() {
  const { isAuthenticated } = useAuthStore()
  const [searchParams, setSearchParams] = useSearchParams()

  const [entries, setEntries] = useState<ExploreEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const activeClass = searchParams.get('class') ?? ''
  const activeSort: Sort = (searchParams.get('sort') as Sort) ?? 'signal'

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    getExplore({ class: activeClass || undefined, sort: activeSort })
      .then(res => setEntries(res.entries))
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [activeClass, activeSort])

  useEffect(() => { load() }, [load])

  function setFilter(key: string, value: string) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      if (value) next.set(key, value)
      else next.delete(key)
      return next
    }, { replace: true })
  }

  function toggleClass(cls: string) {
    setFilter('class', activeClass === cls ? '' : cls)
  }

  return (
    <div className="min-h-screen bg-void-950 text-ink-50">

      {/* Nav — matches Landing */}
      <header
        className="fixed top-0 inset-x-0 z-50 px-6"
        style={{
          borderBottom: '1px solid rgba(74,64,50,0.4)',
          background: 'rgba(19,14,7,0.88)',
          backdropFilter: 'blur(12px)',
          WebkitBackdropFilter: 'blur(12px)',
        }}
      >
        <div className="max-w-6xl mx-auto h-14 flex items-center justify-between">
          <Link
            to="/"
            className="text-gold-400 text-xs tracking-[0.4em] uppercase hover:text-gold-300 transition-colors"
            style={FANTASY_FONT}
          >
            ◈ Signal
          </Link>
          <nav className="flex items-center gap-5">
            {isAuthenticated() ? (
              <Link
                to="/hub"
                className="text-xs px-4 py-2 text-ink-200 hover:text-gold-300 transition-colors tracking-widest uppercase"
                style={{ border: '1px solid rgba(100,80,30,0.45)' }}
              >
                My Hub →
              </Link>
            ) : (
              <>
                <Link to="/login" className="text-xs text-ink-400 hover:text-ink-200 transition-colors hidden sm:block tracking-widest uppercase">
                  Sign in
                </Link>
                <Link
                  to="/join"
                  className="text-xs px-4 py-2 text-ink-200 hover:text-gold-300 transition-colors tracking-widest uppercase"
                  style={{ border: '1px solid rgba(100,80,30,0.45)' }}
                >
                  Join →
                </Link>
              </>
            )}
          </nav>
        </div>
      </header>

      {/* Page header */}
      <div
        className="pt-14"
        style={{ borderBottom: '1px solid rgba(74,64,50,0.35)' }}
      >
        <div className="max-w-6xl mx-auto px-6 py-10 flex flex-col sm:flex-row sm:items-end justify-between gap-6">
          <div>
            <p
              className="text-[10px] tracking-[0.5em] uppercase mb-3"
              style={{ ...FANTASY_FONT, color: 'rgba(160,112,24,0.65)' }}
            >
              Leaderboard
            </p>
            <h1 className="text-3xl font-bold text-ink-50 tracking-tight" style={FANTASY_FONT}>
              Explore Hunters
            </h1>
            {!loading && !error && (
              <p className="text-xs text-ink-600 mt-2 tracking-wide">
                {entries.length > 0
                  ? `${entries.length} published signal profile${entries.length !== 1 ? 's' : ''}`
                  : 'No profiles match this filter'}
              </p>
            )}
          </div>

          {/* Sort toggle */}
          <div
            className="flex items-center"
            style={{ border: '1px solid rgba(74,64,50,0.45)' }}
          >
            {(['signal', 'recent'] as Sort[]).map(s => (
              <button
                key={s}
                onClick={() => setFilter('sort', s)}
                className="text-xs px-5 py-2 tracking-widest uppercase transition-colors"
                style={{
                  color: activeSort === s ? '#c99224' : '#4a4032',
                  background: activeSort === s ? 'rgba(160,112,24,0.08)' : 'transparent',
                  borderRight: s === 'signal' ? '1px solid rgba(74,64,50,0.45)' : undefined,
                  ...FANTASY_FONT,
                }}
              >
                {s === 'signal' ? 'By Signal' : 'Recent'}
              </button>
            ))}
          </div>
        </div>

        {/* Class filter pills */}
        <div className="max-w-6xl mx-auto px-6 pb-5 flex flex-wrap gap-2">
          <button
            onClick={() => setFilter('class', '')}
            className="text-xs px-4 py-2 tracking-widest uppercase transition-colors"
            style={{
              border: !activeClass ? '1px solid rgba(160,112,24,0.6)' : '1px solid rgba(74,64,50,0.45)',
              color: !activeClass ? '#c99224' : '#4a4032',
              background: !activeClass ? 'rgba(160,112,24,0.06)' : 'transparent',
            }}
          >
            All
          </button>
          {ALL_CLASSES.map(cls => (
            <button
              key={cls}
              onClick={() => toggleClass(cls)}
              className="text-xs px-4 py-2 tracking-widest uppercase transition-colors inline-flex items-center gap-2"
              style={{
                border: activeClass === cls ? '1px solid rgba(160,112,24,0.6)' : '1px solid rgba(74,64,50,0.45)',
                color: activeClass === cls ? '#c99224' : '#4a4032',
                background: activeClass === cls ? 'rgba(160,112,24,0.06)' : 'transparent',
              }}
            >
              <span style={{ color: activeClass === cls ? '#c99224' : '#3a2c15' }}>
                <ClassIcon cls={cls} size={10} />
              </span>
              {cls.replace('The ', '')}
            </button>
          ))}
        </div>
      </div>

      {/* Content */}
      <main className="max-w-6xl mx-auto px-6 py-10">
        {error && (
          <div className="p-12 text-center" style={{ border: '1px solid rgba(74,64,50,0.45)' }}>
            <p className="text-ink-400 text-sm mb-4">Failed to load profiles.</p>
            <button
              onClick={load}
              className="text-xs tracking-widest uppercase px-5 py-2 transition-colors hover:text-gold-300"
              style={{ border: '1px solid rgba(74,64,50,0.5)', color: '#4a4032' }}
            >
              Retry
            </button>
          </div>
        )}

        {loading && (
          <div
            className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3"
            style={{ gap: 1, background: 'rgba(74,64,50,0.3)', border: '1px solid rgba(74,64,50,0.45)' }}
          >
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="p-6 space-y-4 animate-pulse" style={{ background: '#130e07' }}>
                <div className="h-5 w-1/3 rounded-none" style={{ background: 'rgba(74,64,50,0.4)' }} />
                <div className="h-7 w-2/3 rounded-none" style={{ background: 'rgba(74,64,50,0.4)' }} />
                <div className="h-3 w-full rounded-none" style={{ background: 'rgba(74,64,50,0.3)' }} />
                <div className="h-3 w-4/5 rounded-none" style={{ background: 'rgba(74,64,50,0.3)' }} />
              </div>
            ))}
          </div>
        )}

        {!loading && !error && entries.length === 0 && (
          <div className="py-20 text-center" style={{ border: '1px solid rgba(74,64,50,0.45)' }}>
            <p
              className="text-[10px] tracking-[0.5em] uppercase mb-4"
              style={{ ...FANTASY_FONT, color: 'rgba(160,112,24,0.65)' }}
            >
              No profiles yet
            </p>
            <p className="text-sm text-ink-400 max-w-sm mx-auto mb-8 leading-relaxed">
              {activeClass
                ? `No published ${activeClass} profiles yet. Be the first.`
                : 'No published profiles yet. Be the first to build and share yours.'}
            </p>
            <Link
              to="/join"
              className="inline-flex items-center text-sm font-medium px-8 py-3 transition-opacity hover:opacity-90"
              style={{
                ...FANTASY_FONT,
                background: 'linear-gradient(135deg, #c99224 0%, #8c6018 100%)',
                color: '#130e07',
                letterSpacing: '0.05em',
              }}
            >
              Build your profile →
            </Link>
          </div>
        )}

        {!loading && !error && entries.length > 0 && (
          <div
            className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3"
            style={{ gap: 1, background: 'rgba(74,64,50,0.3)', border: '1px solid rgba(74,64,50,0.45)' }}
          >
            {entries.map(entry => (
              <ProfileCard key={entry.user_id} entry={entry} />
            ))}
          </div>
        )}
      </main>

      {!loading && !error && entries.length > 0 && (
        <footer
          className="px-6 py-10 text-center"
          style={{ borderTop: '1px solid rgba(74,64,50,0.3)' }}
        >
          <p className="text-xs text-ink-600 mb-5 tracking-wide">
            Signal profiles are artifact-driven — every score is backed by evidence.
          </p>
          <Link
            to="/join"
            className="inline-flex items-center text-xs font-medium px-7 py-2.5 transition-opacity hover:opacity-90"
            style={{
              ...FANTASY_FONT,
              background: 'linear-gradient(135deg, #c99224 0%, #8c6018 100%)',
              color: '#130e07',
              letterSpacing: '0.04em',
            }}
          >
            Build your own profile →
          </Link>
        </footer>
      )}
    </div>
  )
}

// ─── Profile card ─────────────────────────────────────────────────────────────

function ProfileCard({ entry }: { entry: ExploreEntry }) {
  const hasSignal = (entry.trust ?? 0) > 0
  const trustPct = Math.round((entry.trust ?? 0) * 100)
  const languages = (entry.top_languages ?? []).slice(0, 3)
  const cls = entry.class ?? ''

  return (
    <Link
      to={`/p/${entry.user_id}`}
      className="group flex flex-col gap-4 p-6 transition-colors"
      style={{ background: '#130e07' }}
      onMouseEnter={e => (e.currentTarget.style.background = '#1c1409')}
      onMouseLeave={e => (e.currentTarget.style.background = '#130e07')}
    >
      {/* Class row */}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex items-start gap-2.5">
          {cls && (
            <span className="mt-0.5 flex-shrink-0" style={{ color: '#c99224' }}>
              <ClassIcon cls={cls} size={14} />
            </span>
          )}
          <div className="min-w-0">
            <p
              className="text-[10px] tracking-[0.35em] uppercase truncate"
              style={{ ...FANTASY_FONT, color: '#c99224' }}
            >
              {cls.replace('The ', '') || '—'}
            </p>
            {entry.subclass && (
              <p className="text-[10px] text-ink-600 tracking-wide truncate mt-0.5">
                {entry.subclass}
              </p>
            )}
          </div>
        </div>

        {hasSignal && (
          <div className="flex-shrink-0 text-right">
            <span className="text-lg font-bold tabular-nums" style={{ color: '#e8c040' }}>
              {trustPct}%
            </span>
            <p className="text-[9px] tracking-widest uppercase" style={{ color: '#3a2c15' }}>Trust</p>
          </div>
        )}
      </div>

      {/* Trust bar */}
      {hasSignal && (
        <div
          className="h-px w-full"
          style={{ background: 'rgba(74,64,50,0.4)' }}
        >
          <div
            className="h-full"
            style={{
              width: `${trustPct}%`,
              background: 'linear-gradient(to right, #8c6018, #c99224)',
            }}
          />
        </div>
      )}

      {/* Headline */}
      {entry.headline && (
        <p className="text-sm text-ink-400 leading-snug line-clamp-2 italic">
          &ldquo;{entry.headline}&rdquo;
        </p>
      )}

      {/* Footer */}
      <div
        className="flex items-center justify-between gap-2 mt-auto pt-3"
        style={{ borderTop: '1px solid rgba(74,64,50,0.3)' }}
      >
        <div className="flex flex-wrap gap-1.5">
          {languages.map(lang => (
            <span
              key={lang}
              className="text-[10px] px-2 py-0.5 text-ink-600"
              style={{ border: '1px solid rgba(74,64,50,0.4)' }}
            >
              {lang}
            </span>
          ))}
        </div>
        {entry.github_username && (
          <span className="text-[10px] text-ink-600 font-mono flex-shrink-0">
            @{entry.github_username}
          </span>
        )}
      </div>
    </Link>
  )
}
