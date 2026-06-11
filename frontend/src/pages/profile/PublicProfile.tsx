import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { getPublicProfile } from '../../api/profile'
import { OrnatePanel } from '../../components/ui/OrnatePanel'
import { DimensionBars } from '../../components/signal/DimensionBars'
import { EmblemSVG } from '../../components/emblem/Emblem'
import type { PublicProfileResponse, UserSignalScore } from '../../types'

const ZERO_SIGNAL: UserSignalScore = {
  user_id: 0,
  output_percentile: 0,
  craft_percentile: 0,
  influence_percentile: 0,
  collaboration_percentile: 0,
  range_percentile: 0,
  trust: 0,
  updated_at: '',
}

// ─── Class icon ───────────────────────────────────────────────────────────────

function ClassIcon({ cls }: { cls: string | null }) {
  const paths: Record<string, string> = {
    'The Architect':  'M10 2L2 8v10h16V8L10 2zM10 5l6 4.5V16H4V9.5L10 5zM8 12h4v4H8z',
    'The Artisan':    'M14 3l3 3-9 9-4 1 1-4 9-9zM5 15l-1 .5.5-1',
    'The Pathfinder': 'M10 2a8 8 0 100 16A8 8 0 0010 2zm0 2l.5 5.5L15 10l-4.5.5L10 16l-.5-5.5L5 10l4.5-.5z',
    'The Sage':       'M6 3h8v1H6zM5 5h10v9H5zM7 7h2v2H7zM11 7h2v2h-2zM7 11h6v1H7z',
    'The Operator':   'M10 7a3 3 0 100 6 3 3 0 000-6zM9 2h2v4H9zM9 14h2v4H9zM2 9h4v2H2zM14 9h4v2h-4z',
    'The Sentinel':   'M10 2L3 6v5c0 4 3 7.5 7 8.5 4-1 7-4.5 7-8.5V6L10 2zm0 2.3l5 2.7V11c0 2.8-2 5.2-5 6.2-3-1-5-3.4-5-6.2V7z',
    'The Artificer':  'M6 4h8v2l2 2v6l-2 2H6l-2-2V8l2-2V4zm2 2v1h4V6H8zm-2 3v4h8V9H6zm2 1h2v2H8v-2z',
  }
  const d = paths[cls ?? ''] ?? paths['The Pathfinder']
  return (
    <svg width="28" height="28" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <path d={d} fill="currentColor" fillRule="evenodd" />
    </svg>
  )
}


// ─── PublicProfile ────────────────────────────────────────────────────────────

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
        <div className="text-center space-y-3">
          <p
            className="text-gold-400 text-sm tracking-[0.4em] uppercase animate-pulse-gold"
            style={{ fontFamily: '"Cinzel", serif' }}
          >
            Consulting the Oracle…
          </p>
          <p className="text-xs tracking-widest" style={{ color: '#4a4032' }}>Reading the signal</p>
        </div>
      </div>
    )
  }

  if (notFound || !data) {
    return (
      <div className="min-h-screen bg-void-950 flex flex-col items-center justify-center gap-6 px-6">
        <p
          className="text-gold-400 text-xs tracking-[0.4em] uppercase"
          style={{ fontFamily: '"Cinzel", serif' }}
        >
          No record found, Hunter
        </p>
        <p className="text-ink-400 text-sm text-center max-w-xs">
          This profile doesn&rsquo;t exist or hasn&rsquo;t been published yet.
        </p>
        <Link to="/" className="text-xs text-ink-600 hover:text-ink-400 transition-colors">
          ← Return home
        </Link>
      </div>
    )
  }

  const { profile, signal, github } = data
  const scores = signal ?? ZERO_SIGNAL
  const isLowTrust = (scores.trust ?? 0) < 0.3

  return (
    <div className="min-h-screen bg-void-950 text-ink-50">

      {/* ── Header ─────────────────────────────────────────────────────── */}
      <header
        className="px-6 py-4 flex items-center justify-between"
        style={{ borderBottom: '1px solid rgba(74,64,50,0.45)' }}
      >
        <Link
          to="/"
          className="text-gold-400 text-xs tracking-[0.4em] uppercase hover:text-gold-300 transition-colors"
          style={{ fontFamily: '"Cinzel", serif' }}
        >
          ◈ Signal
        </Link>
        <div className="flex items-center gap-5">
          <Link
            to="/explore"
            className="text-xs text-ink-400 hover:text-ink-200 transition-colors tracking-widest uppercase"
          >
            Explore
          </Link>
          <Link
            to="/join"
            className="text-xs px-4 py-2 text-ink-200 hover:text-gold-300 transition-colors tracking-widest uppercase"
            style={{ border: '1px solid rgba(100,80,30,0.45)' }}
          >
            Build yours →
          </Link>
        </div>
      </header>

      {/* ── Hero ───────────────────────────────────────────────────────── */}
      <section
        className="px-6 py-14 relative overflow-hidden"
        style={{ borderBottom: '1px solid rgba(74,64,50,0.35)' }}
      >
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            background: 'radial-gradient(ellipse 55% 65% at 50% 85%, rgba(180,130,30,0.06) 0%, transparent 70%)',
          }}
        />

        <div className="relative max-w-4xl mx-auto">
          <div className="flex flex-col md:flex-row items-center md:items-start gap-10">

            <div className="flex-shrink-0">
              <EmblemSVG score={scores} cls={profile.class ?? 'The Pathfinder'} size={190} />
            </div>

            <div className="flex-1 text-center md:text-left">
              <p
                className="text-[10px] tracking-[0.5em] uppercase mb-5"
                style={{ fontFamily: '"Cinzel", serif', color: 'rgba(160,112,24,0.65)' }}
              >
                Signal Profile
              </p>

              {/* Class */}
              <div className="flex items-center gap-3 justify-center md:justify-start mb-1.5">
                <span style={{ color: '#c99224' }}>
                  <ClassIcon cls={profile.class} />
                </span>
                <h1
                  className="leading-none"
                  style={{
                    fontFamily: '"Cinzel", serif',
                    fontWeight: 700,
                    fontSize: 'clamp(2.2rem, 6vw, 3.5rem)',
                    color: '#e8c040',
                    textShadow: '0 0 40px rgba(200,148,40,0.18)',
                  }}
                >
                  {profile.class ?? '—'}
                </h1>
              </div>

              {/* Subclass */}
              <p
                className="text-sm tracking-[0.3em] uppercase mb-5"
                style={{ fontFamily: '"Cinzel", serif', color: '#b48638' }}
              >
                {profile.subclass ?? ''}
              </p>

              {/* Ornamental divider */}
              <div className="flex items-center gap-3 justify-center md:justify-start mb-5">
                <div className="h-px w-14" style={{ background: 'rgba(160,112,24,0.3)' }} />
                <span style={{ color: 'rgba(160,112,24,0.35)', fontSize: 7 }}>◆</span>
                <div className="h-px w-8" style={{ background: 'rgba(160,112,24,0.3)' }} />
              </div>

              {/* Headline */}
              {profile.headline && (
                <p className="text-base text-ink-200 italic leading-relaxed mb-6 max-w-xl">
                  &ldquo;{profile.headline}&rdquo;
                </p>
              )}

              {/* Badges */}
              <div className="flex flex-wrap items-center gap-3 justify-center md:justify-start">
                {github && (
                  <a
                    href={`https://github.com/${github.github_username}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-2 text-xs hover:text-ink-200 transition-colors"
                    style={{
                      border: '1px solid rgba(74,64,50,0.55)',
                      padding: '5px 11px',
                      color: '#8a7e6a',
                      fontFamily: 'ui-monospace, monospace',
                    }}
                  >
                    @{github.github_username}
                    <span style={{ color: 'rgba(90,70,20,0.5)' }}>·</span>
                    <span style={{ color: '#5ab870', letterSpacing: '0.1em', fontSize: 9, textTransform: 'uppercase' }}>
                      ✓ Verified
                    </span>
                  </a>
                )}
                {isLowTrust && (
                  <span
                    className="text-[9px] tracking-widest uppercase px-2 py-1"
                    style={{ border: '1px solid rgba(180,134,56,0.28)', color: '#8c6424' }}
                  >
                    Limited signal
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── Main content ───────────────────────────────────────────────── */}
      <main className="max-w-4xl mx-auto px-6 py-12 space-y-10">

        <div className="grid lg:grid-cols-2 gap-10 items-start">
          {/* Attributes */}
          <OrnatePanel title="Attributes">
            <div className="px-6 py-5">
              <DimensionBars scores={scores} />
            </div>
          </OrnatePanel>

          {/* About */}
          <OrnatePanel title="About">
            <div className="px-6 py-5 space-y-7">
              {profile.summary && (
                <p className="text-sm text-ink-200 leading-relaxed">{profile.summary}</p>
              )}

              {(profile.strengths ?? []).length > 0 && (
                <div>
                  <p
                    className="text-[10px] tracking-[0.3em] uppercase mb-3"
                    style={{ fontFamily: '"Cinzel", serif', color: '#c99224' }}
                  >
                    Strengths
                  </p>
                  <ul className="space-y-2">
                    {profile.strengths.map((s, i) => (
                      <li key={i} className="flex items-start gap-2.5 text-sm text-ink-200">
                        <span className="flex-shrink-0 mt-1" style={{ color: '#b48638', fontSize: 8 }}>◆</span>
                        {s}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {(profile.growth_paths ?? []).length > 0 && (
                <div>
                  <p
                    className="text-[10px] tracking-[0.3em] uppercase mb-3"
                    style={{ fontFamily: '"Cinzel", serif', color: '#7a5210' }}
                  >
                    Growth Path
                  </p>
                  <ul className="space-y-2.5">
                    {profile.growth_paths.map((g, i) => (
                      <li key={i} className="flex items-start gap-2.5 text-sm text-ink-400">
                        <span className="flex-shrink-0 mt-0.5" style={{ color: '#4a4032' }}>→</span>
                        {g}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </OrnatePanel>
        </div>

        {/* GitHub evidence */}
        {github && (
          <div
            className="px-6 py-5"
            style={{ border: '1px solid rgba(74,64,50,0.4)', background: 'rgba(22,16,7,0.55)' }}
          >
            <div className="flex items-center justify-between mb-5">
              <div className="flex items-center gap-3">
                <span
                  className="text-xs tracking-[0.25em] uppercase"
                  style={{ fontFamily: '"Cinzel", serif', color: '#c99224' }}
                >
                  GitHub Evidence
                </span>
                <span
                  className="text-[9px] tracking-widest uppercase px-2 py-0.5"
                  style={{ border: '1px solid rgba(90,184,112,0.3)', color: '#5ab870' }}
                >
                  Platform Verified
                </span>
              </div>
              <a
                href={`https://github.com/${github.github_username}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs text-ink-500 hover:text-ink-300 transition-colors"
                style={{ fontFamily: 'ui-monospace, monospace' }}
              >
                @{github.github_username} ↗
              </a>
            </div>

            <div className="flex gap-10 mb-4">
              {[
                { label: 'Repos',     value: github.repo_count },
                { label: 'Stars',     value: github.star_count },
                { label: 'Followers', value: github.followers },
              ].map(s => (
                <div key={s.label} className="flex flex-col gap-0.5">
                  <span className="text-2xl font-bold tabular-nums text-ink-50">{s.value}</span>
                  <span className="text-[9px] tracking-widest uppercase" style={{ color: '#4a4032' }}>{s.label}</span>
                </div>
              ))}
            </div>

            {github.top_languages.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {github.top_languages.map(lang => (
                  <span
                    key={lang}
                    className="text-xs px-2.5 py-1"
                    style={{ border: '1px solid rgba(74,64,50,0.45)', color: '#8a7e6a', background: 'rgba(42,30,14,0.4)' }}
                  >
                    {lang}
                  </span>
                ))}
              </div>
            )}
          </div>
        )}
      </main>

      {/* ── Footer CTA ─────────────────────────────────────────────────── */}
      <footer
        className="px-6 py-14 text-center"
        style={{ borderTop: '1px solid rgba(74,64,50,0.35)' }}
      >
        <p
          className="text-[10px] tracking-[0.5em] uppercase mb-2"
          style={{ fontFamily: '"Cinzel", serif', color: 'rgba(160,112,24,0.55)' }}
        >
          Multi-dimensional developer signal
        </p>
        <p className="text-sm text-ink-400 mb-8 max-w-sm mx-auto">
          Build your own verified signal profile. Connect your evidence, claim your class.
        </p>
        <Link
          to="/join"
          className="inline-flex items-center gap-2 text-sm font-medium px-8 py-3 transition-opacity hover:opacity-90"
          style={{
            background: 'linear-gradient(135deg, #c99224 0%, #8c6018 100%)',
            color: '#130e07',
            fontFamily: '"Cinzel", serif',
            letterSpacing: '0.05em',
          }}
        >
          Forge your identity →
        </Link>
        <p className="mt-8 text-[9px] tracking-widest" style={{ color: 'rgba(74,64,50,0.55)' }}>
          Powered by Signal
        </p>
      </footer>

    </div>
  )
}
