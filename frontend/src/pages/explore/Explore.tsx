import { useEffect, useState, useCallback } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { getExplore } from '../../api/explore'
import { useAuthStore } from '../../store/auth'
import type { ExploreEntry } from '../../types'

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

      {/* Header */}
      <header className="border-b border-void-800 px-6 py-4 flex items-center justify-between">
        <Link to="/" className="text-gold-400 text-xs tracking-[0.35em] uppercase hover:text-gold-300 transition-colors">
          ◈ Signal
        </Link>
        <nav className="flex items-center gap-4">
          {isAuthenticated() ? (
            <Link to="/hub" className="text-xs text-ink-400 hover:text-ink-200 transition-colors">
              Hub →
            </Link>
          ) : (
            <>
              <Link to="/login" className="text-xs text-ink-400 hover:text-ink-200 transition-colors">
                Sign in
              </Link>
              <Link
                to="/join"
                className="text-xs border border-void-600 px-4 py-2 text-ink-300 hover:border-gold-400 hover:text-gold-300 transition-colors"
              >
                Build yours →
              </Link>
            </>
          )}
        </nav>
      </header>

      {/* Page title + sort */}
      <div className="border-b border-void-800 px-6 py-6 flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-ink-50 tracking-tight">Explore Profiles</h1>
          <p className="text-xs text-ink-600 mt-1">
            {entries.length > 0
              ? `${entries.length} published signal profile${entries.length !== 1 ? 's' : ''}`
              : loading ? '' : 'No profiles match this filter'}
          </p>
        </div>

        {/* Sort toggle */}
        <div className="flex items-center gap-1 border border-void-700 p-0.5">
          {(['signal', 'recent'] as Sort[]).map(s => (
            <button
              key={s}
              onClick={() => setFilter('sort', s)}
              className={`text-xs px-4 py-1.5 tracking-wide transition-colors ${
                activeSort === s
                  ? 'bg-void-700 text-ink-100'
                  : 'text-ink-500 hover:text-ink-300'
              }`}
            >
              {s === 'signal' ? 'By Signal' : 'Recent'}
            </button>
          ))}
        </div>
      </div>

      {/* Class filter pills */}
      <div className="border-b border-void-800 px-6 py-4 flex flex-wrap gap-2">
        <button
          onClick={() => setFilter('class', '')}
          className={`text-xs px-3 py-1.5 border tracking-wide transition-colors ${
            !activeClass
              ? 'border-gold-400 text-gold-400 bg-gold-400/5'
              : 'border-void-600 text-ink-500 hover:border-void-500 hover:text-ink-300'
          }`}
        >
          All classes
        </button>
        {ALL_CLASSES.map(cls => (
          <button
            key={cls}
            onClick={() => toggleClass(cls)}
            className={`text-xs px-3 py-1.5 border tracking-wide transition-colors ${
              activeClass === cls
                ? 'border-gold-400 text-gold-400 bg-gold-400/5'
                : 'border-void-600 text-ink-500 hover:border-void-500 hover:text-ink-300'
            }`}
          >
            {cls.replace('The ', '')}
          </button>
        ))}
      </div>

      {/* Content */}
      <main className="max-w-6xl mx-auto px-6 py-10">
        {error && (
          <div className="border border-void-700 p-8 text-center">
            <p className="text-ink-500 text-sm">Failed to load profiles.</p>
            <button onClick={load} className="mt-3 text-xs text-gold-400 hover:text-gold-300">
              Retry
            </button>
          </div>
        )}

        {loading && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-px bg-void-700 border border-void-700">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="bg-void-950 p-6 space-y-3 animate-pulse">
                <div className="h-7 w-3/4 bg-void-800 rounded-none" />
                <div className="h-3 w-1/2 bg-void-800 rounded-none" />
                <div className="h-4 w-full bg-void-800 rounded-none mt-4" />
                <div className="h-4 w-4/5 bg-void-800 rounded-none" />
              </div>
            ))}
          </div>
        )}

        {!loading && !error && entries.length === 0 && (
          <div className="border border-void-700 p-16 text-center">
            <p className="text-gold-400 text-xs tracking-[0.3em] uppercase mb-3">No profiles yet</p>
            <p className="text-sm text-ink-500 max-w-sm mx-auto mb-6">
              {activeClass
                ? `No published ${activeClass} profiles yet. Be the first.`
                : 'No published profiles yet. Be the first to build and share yours.'}
            </p>
            <Link
              to="/join"
              className="inline-flex items-center bg-gold-400 text-void-950 text-xs font-medium px-6 py-2.5 hover:bg-gold-300 transition-colors"
            >
              Build your profile →
            </Link>
          </div>
        )}

        {!loading && !error && entries.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-px bg-void-700 border border-void-700">
            {entries.map(entry => (
              <ProfileCard key={entry.user_id} entry={entry} />
            ))}
          </div>
        )}
      </main>

      {/* Footer */}
      {!loading && !error && entries.length > 0 && (
        <footer className="border-t border-void-800 px-6 py-8 text-center">
          <p className="text-xs text-ink-600 mb-4">
            Signal profiles are artifact-driven — every score is backed by evidence.
          </p>
          <Link
            to="/join"
            className="inline-flex items-center border border-void-600 text-xs px-6 py-2.5 text-ink-300 hover:border-gold-400 hover:text-gold-300 transition-colors"
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
  const hasSignal = entry.total_signal > 0
  const languages = (entry.top_languages ?? []).slice(0, 3)

  return (
    <Link
      to={`/p/${entry.user_id}`}
      className="group bg-void-950 p-6 flex flex-col gap-4 hover:bg-void-900 transition-colors"
    >
      {/* Class + signal */}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-xl font-black tracking-tight text-gold-400 leading-none truncate group-hover:text-gold-300 transition-colors">
            {entry.class ?? '—'}
          </h2>
          <p className="mt-1.5 text-[10px] text-ink-500 tracking-[0.2em] uppercase">
            {entry.subclass ?? ''}
          </p>
        </div>
        {hasSignal && (
          <div className="flex-shrink-0 text-right">
            <span className="text-xl font-bold text-ink-200 tabular-nums">{entry.total_signal}</span>
            <p className="text-[9px] text-ink-600 tracking-widest uppercase">Signal</p>
          </div>
        )}
      </div>

      {/* Headline */}
      {entry.headline && (
        <p className="text-sm text-ink-400 leading-snug line-clamp-2 italic">
          &ldquo;{entry.headline}&rdquo;
        </p>
      )}

      {/* Footer: languages + github */}
      <div className="flex items-center justify-between gap-2 mt-auto pt-2 border-t border-void-800">
        <div className="flex flex-wrap gap-1.5">
          {languages.map(lang => (
            <span key={lang} className="text-[10px] px-2 py-0.5 border border-void-700 text-ink-500 bg-void-900">
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
