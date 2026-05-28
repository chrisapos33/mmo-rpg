import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { generateBuild, getBuild } from '../../api/onboarding'
import { ApiError } from '../../api/client'
import { Button } from '../../components/ui/Button'
import type { Profile } from '../../types'

// ─── Forge messages shown while Claude generates the build ────────────────────

const FORGE_MESSAGES = [
  'Analyzing your professional signal...',
  'Mapping your career architecture...',
  'Identifying your class...',
  'Assembling your build identity...',
  'Calibrating specialization...',
  'Forging your profile...',
]

// ─── Component ────────────────────────────────────────────────────────────────

export function Reveal() {
  const navigate = useNavigate()
  const [stage, setStage] = useState<'forging' | 'revealed' | 'error'>('forging')
  const [profile, setProfile] = useState<Profile | null>(null)
  const [errorMsg, setErrorMsg] = useState('')
  const [forgeIdx, setForgeIdx] = useState(0)

  // Cycle forge messages
  useEffect(() => {
    if (stage !== 'forging') return
    const id = setInterval(() => setForgeIdx(i => (i + 1) % FORGE_MESSAGES.length), 2400)
    return () => clearInterval(id)
  }, [stage])

  // On mount: check for existing build, or generate one
  useEffect(() => {
    const MIN_FORGE_MS = 2000 // always show forge state for at least this long
    const start = Date.now()

    async function init() {
      try {
        let p: Profile
        try {
          p = await getBuild()
        } catch (err) {
          if (err instanceof ApiError && err.status === 404) {
            p = await generateBuild()
          } else {
            throw err
          }
        }

        const elapsed = Date.now() - start
        if (elapsed < MIN_FORGE_MS) {
          await delay(MIN_FORGE_MS - elapsed)
        }

        setProfile(p)
        setStage('revealed')
      } catch (err) {
        setErrorMsg(err instanceof Error ? err.message : 'Build generation failed.')
        setStage('error')
      }
    }

    init()
  }, [])

  if (stage === 'error') {
    return (
      <div className="min-h-screen bg-void-950 flex flex-col items-center justify-center px-6 gap-6">
        <div className="border border-danger-400/40 bg-void-900 p-8 max-w-md w-full">
          <p className="text-sm text-danger-400 mb-1">Build generation failed</p>
          <p className="text-sm text-ink-400">{errorMsg}</p>
        </div>
        <div className="flex gap-4">
          <Button variant="ghost" onClick={() => navigate('/onboarding/upload')}>
            ← Re-upload CV
          </Button>
          <Button onClick={() => { setStage('forging'); setErrorMsg('') }}>
            Try again
          </Button>
        </div>
      </div>
    )
  }

  if (stage === 'forging' || !profile) {
    return <ForgingScreen message={FORGE_MESSAGES[forgeIdx]} />
  }

  return <RevealScreen profile={profile} onConfirm={() => navigate('/hub')} />
}

// ─── Forging screen ───────────────────────────────────────────────────────────

function ForgingScreen({ message }: { message: string }) {
  return (
    <div className="min-h-screen bg-void-950 flex flex-col items-center justify-center px-6">
      <div className="flex flex-col items-center gap-10 text-center">
        {/* Pulsing scanner icon */}
        <div className="relative w-20 h-20">
          <div className="absolute inset-0 border border-gold-400/20 animate-ping" style={{ animationDuration: '2s' }} />
          <div className="absolute inset-3 border border-gold-400/40 animate-ping" style={{ animationDuration: '2s', animationDelay: '0.4s' }} />
          <div className="absolute inset-0 flex items-center justify-center">
            <span className="text-gold-400 text-3xl">◈</span>
          </div>
        </div>

        <div>
          <p className="text-gold-400 text-xs tracking-[0.4em] uppercase mb-3">Identity Engine</p>
          <p className="text-lg text-ink-200 min-h-[1.75rem] transition-all duration-500">{message}</p>
          <p className="text-xs text-ink-600 mt-2">This takes 10–20 seconds</p>
        </div>

        {/* Animated scan line */}
        <div className="w-64 h-px bg-void-700 relative overflow-hidden">
          <div
            className="absolute inset-y-0 w-1/2 bg-gradient-to-r from-transparent via-gold-400/50 to-transparent"
            style={{ animation: 'scan 2.4s ease-in-out infinite' }}
          />
        </div>
      </div>
    </div>
  )
}

// ─── Reveal screen ────────────────────────────────────────────────────────────

function RevealScreen({ profile, onConfirm }: { profile: Profile; onConfirm: () => void }) {
  return (
    <div className="min-h-screen bg-void-950 flex flex-col items-center justify-center px-6 py-20">
      <div className="w-full max-w-2xl">

        {/* Header flash — "CLASS IDENTIFIED" */}
        <div className="animate-fade-up text-center mb-12">
          <p className="text-gold-400 text-xs tracking-[0.5em] uppercase">
            ◈ &nbsp; Class Identified
          </p>
        </div>

        {/* Class + Subclass */}
        <div className="animate-fade-up delay-100 text-center mb-8">
          <h1 className="text-5xl sm:text-6xl md:text-7xl font-black tracking-tight text-gold-400 leading-none">
            {profile.class}
          </h1>
          <p className="mt-3 text-lg text-ink-200 tracking-widest uppercase">
            {profile.subclass}
          </p>
        </div>

        {/* Headline */}
        {profile.headline && (
          <div className="animate-fade-up border-t border-b border-void-700 py-6 text-center mb-10" style={{ animationDelay: '200ms' }}>
            <p className="text-base text-ink-200 italic leading-relaxed max-w-xl mx-auto">
              &ldquo;{profile.headline}&rdquo;
            </p>
          </div>
        )}

        {/* Summary */}
        {profile.summary && (
          <div className="animate-fade-up mb-10" style={{ animationDelay: '300ms' }}>
            <p className="text-sm text-ink-400 leading-relaxed text-center max-w-xl mx-auto">
              {profile.summary}
            </p>
          </div>
        )}

        {/* Strengths + Growth paths */}
        <div className="animate-fade-up grid sm:grid-cols-2 gap-px bg-void-700 border border-void-700 mb-10" style={{ animationDelay: '400ms' }}>
          {/* Strengths */}
          <div className="bg-void-950 p-6">
            <p className="text-xs text-gold-400 tracking-[0.3em] uppercase mb-4">Strengths</p>
            <ul className="flex flex-col gap-2">
              {(profile.strengths ?? []).map((s, i) => (
                <li key={i} className="flex items-start gap-2 text-sm text-ink-200">
                  <span className="text-gold-400 mt-0.5 flex-shrink-0">◆</span>
                  {s}
                </li>
              ))}
            </ul>
          </div>

          {/* Growth paths */}
          <div className="bg-void-950 p-6">
            <p className="text-xs text-gold-400 tracking-[0.3em] uppercase mb-4">Growth Path</p>
            <ul className="flex flex-col gap-2">
              {(profile.growth_paths ?? []).map((g, i) => (
                <li key={i} className="flex items-start gap-2 text-sm text-ink-200">
                  <span className="text-ink-600 mt-0.5 flex-shrink-0">→</span>
                  {g}
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* CTA */}
        <div className="animate-fade-up text-center" style={{ animationDelay: '550ms' }}>
          <Button onClick={onConfirm} className="px-12 py-3 text-sm">
            This is my build →
          </Button>
          <p className="mt-4 text-xs text-ink-600">
            You can edit and refine your profile at any time.
          </p>
        </div>
      </div>
    </div>
  )
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}
