import { Link } from 'react-router-dom'
import { useAuthStore } from '../../store/auth'
import { Button } from '../../components/ui/Button'

// ─── Landing ──────────────────────────────────────────────────────────────────

export function Landing() {
  return (
    <div className="bg-void-950 text-ink-50 min-h-screen">
      <Nav />
      <main>
        <Hero />
        <Pillars />
        <HowItWorks />
        <CompanyCTA />
      </main>
      <Footer />
    </div>
  )
}

// ─── Nav ──────────────────────────────────────────────────────────────────────

function Nav() {
  const isAuthenticated = useAuthStore(s => s.isAuthenticated())

  return (
    <header className="fixed top-0 inset-x-0 z-50 border-b border-void-800/60 bg-void-950/75 backdrop-blur-md">
      <div className="max-w-6xl mx-auto px-6 h-14 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 text-sm font-semibold tracking-widest text-ink-50">
          <span className="text-gold-400">◈</span>
          <span>SIGNAL</span>
        </Link>
        <div className="flex items-center gap-6">
          {isAuthenticated ? (
            <Link to="/hub">
              <Button variant="ghost" className="text-xs px-4 py-2">Enter Hub</Button>
            </Link>
          ) : (
            <>
              <Link
                to="/login"
                className="text-sm text-ink-400 hover:text-ink-200 transition-colors hidden sm:block"
              >
                Sign in
              </Link>
              <Link to="/join">
                <Button className="text-xs px-5 py-2">Join the Hunt</Button>
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
    <section className="relative min-h-screen flex flex-col items-center justify-center px-6 overflow-hidden">

      {/* Subtle gold grid */}
      <div
        className="absolute inset-0"
        style={{
          backgroundImage: [
            'linear-gradient(to right, rgba(212,168,67,0.05) 1px, transparent 1px)',
            'linear-gradient(to bottom, rgba(212,168,67,0.05) 1px, transparent 1px)',
          ].join(','),
          backgroundSize: '80px 80px',
        }}
      />

      {/* Central glow behind headline */}
      <div className="absolute inset-0 pointer-events-none">
        <div
          className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-[900px] h-[500px]"
          style={{
            background: 'radial-gradient(ellipse, rgba(212,168,67,0.07) 0%, transparent 65%)',
          }}
        />
      </div>

      {/* Vignette — kills the grid at edges so it doesn't look tiled */}
      <div
        className="absolute inset-0"
        style={{
          background: 'radial-gradient(ellipse at center, transparent 45%, #06080f 100%)',
        }}
      />

      {/* Content */}
      <div className="relative z-10 max-w-4xl mx-auto text-center">
        <p className="animate-fade-up text-gold-400 text-xs tracking-[0.4em] uppercase mb-8">
          Hunter Identity Platform
        </p>

        <h1 className="animate-fade-up delay-100 text-6xl sm:text-7xl md:text-[88px] font-bold tracking-tight leading-[1.04]">
          Not a resume.
          <br />
          <span className="text-gold-400">A build.</span>
        </h1>

        <p className="animate-fade-up delay-200 mt-8 text-lg text-ink-400 max-w-xl mx-auto leading-relaxed">
          The professional identity platform for engineers and builders.
          Your class, signal, and specializations — derived from what you've actually shipped.
        </p>

        <div className="animate-fade-up delay-300 mt-12 flex flex-col sm:flex-row items-center justify-center gap-4">
          <Link to="/join">
            <Button className="px-8 py-3 text-sm">
              Begin your build →
            </Button>
          </Link>
          <button
            disabled
            className="px-8 py-3 text-sm border border-void-700 text-ink-600 cursor-not-allowed tracking-wide"
          >
            For Companies — Coming Soon
          </button>
        </div>
      </div>

      {/* Scroll indicator */}
      <div className="animate-fade-up delay-500 absolute bottom-10 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2.5">
        <span className="text-[10px] text-ink-600 tracking-[0.3em] uppercase">Scroll</span>
        <div className="w-px h-8 bg-gradient-to-b from-void-600 to-transparent" />
      </div>
    </section>
  )
}

// ─── Pillars ──────────────────────────────────────────────────────────────────

const PILLARS = [
  {
    id: '01',
    label: 'Signal',
    headline: 'Reputation from real work.',
    body: "Your signal is computed from what you've actually built — merged PRs, OSS contributions, verified skills, and professional track record. No self-reported noise.",
  },
  {
    id: '02',
    label: 'Build',
    headline: 'Identity, not a bio.',
    body: 'Connect your GitHub. Upload your CV. AI reads your actual output and assigns your class, subclass, and specialization — then generates your professional build summary.',
  },
  {
    id: '03',
    label: 'Hunt',
    headline: 'Be found. Don\'t apply.',
    body: 'Strong signal attracts the right teams. Companies search by specialization and signal depth — not keyword-stuffed resumes and cold applications.',
  },
]

function Pillars() {
  return (
    <section className="max-w-6xl mx-auto px-6 py-32">
      <div className="mb-16 text-center">
        <p className="text-gold-400 text-xs tracking-[0.4em] uppercase mb-4">The Platform</p>
        <h2 className="text-3xl font-bold text-ink-50 tracking-tight">How it's different</h2>
      </div>

      {/* gap-px + bg-void-700 creates hairline dividers between cells */}
      <div className="border border-void-700">
        <div className="grid md:grid-cols-3 gap-px bg-void-700">
          {PILLARS.map(p => (
            <div
              key={p.id}
              className="bg-void-950 hover:bg-void-900 transition-colors p-10 flex flex-col gap-6 group"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs text-gold-400 tracking-[0.3em] uppercase group-hover:text-gold-300 transition-colors">
                  {p.label}
                </span>
                <span className="text-xs text-void-600 font-mono">{p.id}</span>
              </div>
              <h3 className="text-xl font-semibold text-ink-50 leading-snug">{p.headline}</h3>
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
    title: 'Upload your CV',
    body: 'Drop your CV or resume. Our AI extracts your roles, skills, technologies, and work history — no manual data entry. Messy input, clean output.',
  },
  {
    n: '02',
    title: 'Connect GitHub',
    body: 'Link your GitHub account. We analyze your repositories, programming languages, contribution patterns, and open-source activity to build a true picture of your output.',
  },
  {
    n: '03',
    title: 'Your build is revealed',
    body: 'Your class, subclass, and specializations are assigned. An AI-generated build summary captures your professional identity in a way a resume never could.',
  },
]

function HowItWorks() {
  return (
    <section className="border-t border-void-800 py-32 px-6">
      <div className="max-w-3xl mx-auto">
        <div className="mb-16">
          <p className="text-gold-400 text-xs tracking-[0.4em] uppercase mb-4">Onboarding</p>
          <h2 className="text-3xl font-bold text-ink-50 tracking-tight">How it works</h2>
        </div>

        <div className="flex flex-col">
          {STEPS.map((step, i) => (
            <div key={step.n} className="flex gap-10 relative pb-12 last:pb-0">
              {/* Connector line between steps */}
              {i < STEPS.length - 1 && (
                <div className="absolute left-5 top-10 bottom-0 w-px bg-void-700" />
              )}

              {/* Step number */}
              <div className="relative flex-shrink-0 w-10 h-10 border border-void-600 flex items-center justify-center text-xs text-gold-400 font-mono bg-void-950 z-10">
                {step.n}
              </div>

              {/* Content */}
              <div className="pt-2 pb-2">
                <h3 className="text-lg font-semibold text-ink-50 mb-2">{step.title}</h3>
                <p className="text-sm text-ink-400 leading-relaxed">{step.body}</p>
              </div>
            </div>
          ))}
        </div>

        {/* End CTA */}
        <div className="mt-16 pt-12 border-t border-void-800">
          <Link to="/join">
            <Button className="px-8 py-3 text-sm">Begin your build →</Button>
          </Link>
        </div>
      </div>
    </section>
  )
}

// ─── Company CTA ──────────────────────────────────────────────────────────────

function CompanyCTA() {
  return (
    <section className="border-t border-void-800 py-32 px-6">
      <div className="max-w-2xl mx-auto border border-void-700 p-12">
        <p className="text-gold-400 text-xs tracking-[0.4em] uppercase mb-6">For Companies</p>
        <h2 className="text-3xl font-bold text-ink-50 tracking-tight mb-4">
          Building a team?
        </h2>
        <p className="text-ink-400 leading-relaxed mb-10">
          We're building the employer side of this. Search by signal depth, specialization,
          and contribution history — not CV keywords. Get notified when it's ready.
        </p>
        <div className="flex gap-3">
          <input
            type="email"
            placeholder="your@company.com"
            className="flex-1 min-w-0 bg-void-800 border border-void-600 px-4 py-2.5 text-sm text-ink-50 placeholder:text-ink-600 outline-none focus:border-gold-500 transition-colors"
          />
          <button className="flex-shrink-0 px-5 py-2.5 text-sm border border-void-600 text-ink-400 hover:border-gold-500 hover:text-ink-200 transition-all bg-void-800 whitespace-nowrap">
            Get early access
          </button>
        </div>
      </div>
    </section>
  )
}

// ─── Footer ───────────────────────────────────────────────────────────────────

function Footer() {
  return (
    <footer className="border-t border-void-800 py-10 px-6">
      <div className="max-w-6xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-6">
        <span className="text-sm font-semibold tracking-widest">
          <span className="text-gold-400">◈</span>{' '}
          <span className="text-ink-50">SIGNAL</span>
        </span>
        <span className="text-xs text-ink-600">
          © {new Date().getFullYear()} — Professional identity for builders
        </span>
        <div className="flex gap-6">
          <Link to="/join" className="text-xs text-ink-400 hover:text-ink-200 transition-colors">
            Join the Hunt
          </Link>
          <Link to="/login" className="text-xs text-ink-400 hover:text-ink-200 transition-colors">
            Sign in
          </Link>
        </div>
      </div>
    </footer>
  )
}
