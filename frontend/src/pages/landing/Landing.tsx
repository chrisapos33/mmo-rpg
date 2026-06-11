import { Link } from 'react-router-dom'
import { useAuthStore } from '../../store/auth'

const FANTASY_FONT = { fontFamily: '"Cinzel", serif' }

// ─── Landing ──────────────────────────────────────────────────────────────────

export function Landing() {
  return (
    <div className="bg-void-950 text-ink-50 min-h-screen">
      <Nav />
      <main>
        <Hero />
        <Pillars />
        <HowItWorks />
      </main>
      <Footer />
    </div>
  )
}

// ─── Nav ──────────────────────────────────────────────────────────────────────

function Nav() {
  const isAuthenticated = useAuthStore(s => s.isAuthenticated())

  return (
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
        <div className="flex items-center gap-5">
          <Link to="/explore" className="text-xs text-ink-400 hover:text-ink-200 transition-colors tracking-widest uppercase hidden sm:block">
            Explore
          </Link>
          {isAuthenticated ? (
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
        </div>
      </div>
    </header>
  )
}

// ─── Hero ─────────────────────────────────────────────────────────────────────

function Hero() {
  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center px-6 overflow-hidden pt-14">
      {/* Warm ambient glow */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          background: [
            'radial-gradient(ellipse 70% 55% at 50% 60%, rgba(180,130,30,0.09) 0%, transparent 65%)',
            'radial-gradient(ellipse 100% 40% at 50% 100%, rgba(140,100,20,0.06) 0%, transparent 60%)',
          ].join(','),
        }}
      />

      {/* Subtle parchment texture overlay */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          backgroundImage: 'linear-gradient(to right, rgba(160,112,24,0.03) 1px, transparent 1px), linear-gradient(to bottom, rgba(160,112,24,0.03) 1px, transparent 1px)',
          backgroundSize: '80px 80px',
        }}
      />
      {/* Vignette */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{ background: 'radial-gradient(ellipse at center, transparent 40%, #130e07 90%)' }}
      />

      <div className="relative z-10 max-w-4xl mx-auto text-center">
        {/* Eyebrow */}
        <p
          className="animate-fade-up text-[10px] tracking-[0.55em] uppercase mb-8"
          style={{ ...FANTASY_FONT, color: 'rgba(160,112,24,0.7)' }}
        >
          Developer Identity Platform
        </p>

        {/* Headline */}
        <h1
          className="animate-fade-up delay-100 leading-tight mb-4"
          style={{
            ...FANTASY_FONT,
            fontWeight: 700,
            fontSize: 'clamp(3rem, 9vw, 6rem)',
            color: '#f2ece0',
          }}
        >
          Not a résumé.
          <br />
          <span style={{ color: '#e8c040' }}>A build.</span>
        </h1>

        {/* Sub-headline */}
        <p className="animate-fade-up delay-200 mt-6 text-base text-ink-400 max-w-lg mx-auto leading-relaxed">
          Connect your GitHub. The scoring engine reads what you&rsquo;ve actually shipped —
          verified by third-party attribution. An AI layer assigns your class, subclass, and
          dimension profile. Your build is the result.
        </p>

        {/* Primary CTA */}
        <div className="animate-fade-up delay-300 mt-12 flex flex-col sm:flex-row items-center justify-center gap-4">
          <Link
            to="/join"
            className="inline-flex items-center gap-2 text-sm font-medium px-9 py-3.5 transition-opacity hover:opacity-90"
            style={{
              ...FANTASY_FONT,
              background: 'linear-gradient(135deg, #c99224 0%, #8c6018 100%)',
              color: '#130e07',
              letterSpacing: '0.06em',
            }}
          >
            Connect GitHub →
          </Link>
          <Link
            to="/explore"
            className="text-sm text-ink-400 hover:text-ink-200 transition-colors tracking-wide"
          >
            Explore profiles
          </Link>
        </div>

        {/* Trust footnote */}
        <p
          className="animate-fade-up delay-500 mt-8 text-[10px] tracking-widest uppercase"
          style={{ color: 'rgba(74,64,50,0.7)' }}
        >
          Scored from real GitHub activity · Not self-reported
        </p>
      </div>

      {/* Scroll indicator */}
      <div className="animate-fade-up delay-500 absolute bottom-10 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2.5">
        <span className="text-[9px] tracking-[0.4em] uppercase" style={{ color: '#4a4032' }}>Scroll</span>
        <div className="w-px h-8" style={{ background: 'linear-gradient(to bottom, rgba(160,112,24,0.35), transparent)' }} />
      </div>
    </section>
  )
}

// ─── Pillars ──────────────────────────────────────────────────────────────────

const PILLARS = [
  {
    n: '01',
    label: 'Signal',
    headline: 'Reputation from real work.',
    body: 'Scored from merged PRs to repos you don\'t own, dependents, review depth, and contribution history. Third-party attribution only — self-controlled signals carry capped weight.',
  },
  {
    n: '02',
    label: 'Build',
    headline: 'A class, not a job title.',
    body: 'Connect GitHub. The engine derives five dimension scores and a Trust meta-score. AI maps the shape to one of seven classes — Architect, Artisan, Pathfinder, Sage, Operator, Sentinel, Artificer.',
  },
  {
    n: '03',
    label: 'Explore',
    headline: 'Be found by the right teams.',
    body: 'A public profile and an Explore grid let companies search by class, stack, and signal depth — not keyword-stuffed résumés. Signal depth is visible; hype is not.',
  },
]

function Pillars() {
  return (
    <section className="max-w-6xl mx-auto px-6 py-32">
      <div className="mb-14 text-center">
        <p
          className="text-[10px] tracking-[0.5em] uppercase mb-4"
          style={{ ...FANTASY_FONT, color: 'rgba(160,112,24,0.65)' }}
        >
          The Platform
        </p>
        <h2 className="text-3xl font-bold text-ink-50 tracking-tight" style={FANTASY_FONT}>
          How it&rsquo;s different
        </h2>
      </div>

      <div style={{ border: '1px solid rgba(74,64,50,0.45)' }}>
        <div className="grid md:grid-cols-3" style={{ gap: 1, background: 'rgba(74,64,50,0.3)' }}>
          {PILLARS.map(p => (
            <div
              key={p.n}
              className="flex flex-col gap-5 p-10 transition-colors"
              style={{ background: '#130e07' }}
              onMouseEnter={e => (e.currentTarget.style.background = '#1c1409')}
              onMouseLeave={e => (e.currentTarget.style.background = '#130e07')}
            >
              <div className="flex items-center justify-between">
                <span
                  className="text-[10px] tracking-[0.35em] uppercase"
                  style={{ ...FANTASY_FONT, color: '#c99224' }}
                >
                  {p.label}
                </span>
                <span className="text-xs font-mono" style={{ color: '#3a2c15' }}>{p.n}</span>
              </div>
              <h3 className="text-lg font-semibold text-ink-50 leading-snug">{p.headline}</h3>
              <p className="text-sm text-ink-400 leading-relaxed">{p.body}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

// ─── How it works ─────────────────────────────────────────────────────────────

const STEPS = [
  {
    n: '01',
    title: 'Create your account',
    body: 'A free account is the starting point. Takes thirty seconds.',
  },
  {
    n: '02',
    title: 'Connect GitHub',
    body: 'OAuth as yourself — we use your token and your rate limit. No credentials stored. Public activity is ingested, scored, and cached.',
  },
  {
    n: '03',
    title: 'Your build is forged',
    body: 'Scoring runs in the background. When it finishes, AI reads the five dimension scores and assigns your class, subclass, and flavor text. Your character profile is live.',
  },
]

function HowItWorks() {
  return (
    <section
      className="py-32 px-6"
      style={{ borderTop: '1px solid rgba(74,64,50,0.3)' }}
    >
      <div className="max-w-3xl mx-auto">
        <div className="mb-14">
          <p
            className="text-[10px] tracking-[0.5em] uppercase mb-4"
            style={{ ...FANTASY_FONT, color: 'rgba(160,112,24,0.65)' }}
          >
            Onboarding
          </p>
          <h2 className="text-3xl font-bold text-ink-50 tracking-tight" style={FANTASY_FONT}>
            Three steps
          </h2>
        </div>

        <div className="flex flex-col">
          {STEPS.map((step, i) => (
            <div key={step.n} className="flex gap-10 relative pb-12 last:pb-0">
              {i < STEPS.length - 1 && (
                <div
                  className="absolute left-5 top-10 bottom-0 w-px"
                  style={{ background: 'rgba(74,64,50,0.5)' }}
                />
              )}
              <div
                className="relative flex-shrink-0 w-10 h-10 flex items-center justify-center text-xs font-mono z-10"
                style={{
                  border: '1px solid rgba(160,112,24,0.4)',
                  background: '#130e07',
                  color: '#c99224',
                  ...FANTASY_FONT,
                }}
              >
                {step.n}
              </div>
              <div className="pt-2">
                <h3 className="text-base font-semibold text-ink-50 mb-2">{step.title}</h3>
                <p className="text-sm text-ink-400 leading-relaxed">{step.body}</p>
              </div>
            </div>
          ))}
        </div>

        <div
          className="mt-16 pt-12 flex items-center gap-6"
          style={{ borderTop: '1px solid rgba(74,64,50,0.3)' }}
        >
          <Link
            to="/join"
            className="inline-flex items-center gap-2 text-sm font-medium px-8 py-3 transition-opacity hover:opacity-90"
            style={{
              ...FANTASY_FONT,
              background: 'linear-gradient(135deg, #c99224 0%, #8c6018 100%)',
              color: '#130e07',
              letterSpacing: '0.05em',
            }}
          >
            Forge your identity →
          </Link>
          <Link to="/explore" className="text-sm text-ink-400 hover:text-ink-200 transition-colors">
            Browse profiles
          </Link>
        </div>
      </div>
    </section>
  )
}

// ─── Footer ───────────────────────────────────────────────────────────────────

function Footer() {
  return (
    <footer
      className="py-10 px-6"
      style={{ borderTop: '1px solid rgba(74,64,50,0.3)' }}
    >
      <div className="max-w-6xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-6">
        <span
          className="text-xs tracking-[0.4em] uppercase text-gold-400"
          style={FANTASY_FONT}
        >
          ◈ Signal
        </span>
        <span className="text-xs" style={{ color: '#3a2c15' }}>
          © {new Date().getFullYear()} — Professional identity for builders
        </span>
        <div className="flex gap-6">
          <Link to="/explore" className="text-xs text-ink-400 hover:text-ink-200 transition-colors tracking-widest uppercase">
            Explore
          </Link>
          <Link to="/join" className="text-xs text-ink-400 hover:text-ink-200 transition-colors tracking-widest uppercase">
            Join
          </Link>
          <Link to="/login" className="text-xs text-ink-400 hover:text-ink-200 transition-colors tracking-widest uppercase">
            Sign in
          </Link>
        </div>
      </div>
    </footer>
  )
}
