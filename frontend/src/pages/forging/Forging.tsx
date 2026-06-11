import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getScoringStatus } from '../../api/github'
import { generateBuild } from '../../api/onboarding'
import { useAuthStore } from '../../store/auth'

// Cosmetic steps — appear progressively while real scoring runs in background.
const COSMETIC_STEPS = [
  'Reading your GitHub repositories',
  'Analyzing contribution patterns',
  'Computing dimension scores',
  'Consulting the Oracle',
]
const STEP_DELAY_MS = [0, 4000, 10000, 18000]

// ─── Rune SVG ─────────────────────────────────────────────────────────────────

function ForgeSigil() {
  return (
    <div className="relative flex items-center justify-center" style={{ width: 220, height: 220 }}>
      {/* Outer ring — slow rotation */}
      <svg
        width="220" height="220" viewBox="0 0 220 220"
        className="absolute animate-rune-spin"
        style={{ opacity: 0.55 }}
      >
        <circle cx="110" cy="110" r="105" stroke="#c99224" strokeWidth="0.8" fill="none" strokeDasharray="4 8" />
        <circle cx="110" cy="110" r="96" stroke="#8c6424" strokeWidth="0.4" fill="none" />
        {/* 8 tick marks */}
        {Array.from({ length: 8 }).map((_, i) => {
          const a = (i * 45 - 90) * (Math.PI / 180)
          const x1 = 110 + 96 * Math.cos(a)
          const y1 = 110 + 96 * Math.sin(a)
          const x2 = 110 + 105 * Math.cos(a)
          const y2 = 110 + 105 * Math.sin(a)
          return <line key={i} x1={x1} y1={y1} x2={x2} y2={y2} stroke="#c99224" strokeWidth="1.5" />
        })}
      </svg>

      {/* Middle ring — reverse rotation */}
      <svg
        width="160" height="160" viewBox="0 0 160 160"
        className="absolute animate-rune-spin-reverse"
        style={{ opacity: 0.45 }}
      >
        <circle cx="80" cy="80" r="74" stroke="#a07018" strokeWidth="0.6" fill="none" />
        {/* 6-pointed star */}
        <polygon
          points="80,12 93,56 138,56 102,82 116,126 80,100 44,126 58,82 22,56 67,56"
          stroke="#c99224" strokeWidth="0.7" fill="none"
        />
      </svg>

      {/* Inner static sigil */}
      <svg
        width="90" height="90" viewBox="0 0 90 90"
        className="absolute"
        style={{ opacity: 0.7 }}
      >
        <circle cx="45" cy="45" r="40" stroke="#c99224" strokeWidth="1" fill="none" />
        <circle cx="45" cy="45" r="12" stroke="#e8c040" strokeWidth="1.2" fill="none" />
        {/* Cardinal marks */}
        <line x1="45" y1="5" x2="45" y2="20" stroke="#e8c040" strokeWidth="1.5" />
        <line x1="45" y1="70" x2="45" y2="85" stroke="#e8c040" strokeWidth="1.5" />
        <line x1="5" y1="45" x2="20" y2="45" stroke="#e8c040" strokeWidth="1.5" />
        <line x1="70" y1="45" x2="85" y2="45" stroke="#e8c040" strokeWidth="1.5" />
        {/* Diagonal marks */}
        <line x1="16" y1="16" x2="26" y2="26" stroke="#c99224" strokeWidth="0.8" />
        <line x1="64" y1="64" x2="74" y2="74" stroke="#c99224" strokeWidth="0.8" />
        <line x1="74" y1="16" x2="64" y2="26" stroke="#c99224" strokeWidth="0.8" />
        <line x1="16" y1="74" x2="26" y2="64" stroke="#c99224" strokeWidth="0.8" />
      </svg>

      {/* Pulsing center dot */}
      <div
        className="absolute rounded-full animate-pulse-gold"
        style={{ width: 8, height: 8, background: '#e8c040' }}
      />
    </div>
  )
}

// ─── Forging ──────────────────────────────────────────────────────────────────

type Phase = 'running' | 'generating' | 'done' | 'failed'

export function Forging() {
  const navigate = useNavigate()
  const user = useAuthStore(s => s.user)
  const [phase, setPhase] = useState<Phase>('running')
  const [visibleSteps, setVisibleSteps] = useState<number>(1)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const stepTimers = useRef<ReturnType<typeof setTimeout>[]>([])

  function clearAll() {
    if (pollRef.current) clearInterval(pollRef.current)
    stepTimers.current.forEach(t => clearTimeout(t))
  }

  async function finishForging() {
    setPhase('generating')
    try {
      await generateBuild()
    } catch {
      // Build may already exist — non-fatal; proceed to profile
    }
    setPhase('done')
    setTimeout(() => {
      if (user) navigate(`/p/${user.id}`, { replace: true })
      else navigate('/hub', { replace: true })
    }, 1200)
  }

  useEffect(() => {
    // Reveal cosmetic steps progressively
    COSMETIC_STEPS.forEach((_, i) => {
      if (i === 0) return // first step shown immediately via initial state
      const t = setTimeout(() => setVisibleSteps(v => Math.max(v, i + 1)), STEP_DELAY_MS[i])
      stepTimers.current.push(t)
    })

    // Poll real scoring status
    let settled = false

    async function poll() {
      if (settled) return
      try {
        const res = await getScoringStatus()
        if (res.status === 'done') {
          settled = true
          clearAll()
          setVisibleSteps(COSMETIC_STEPS.length)
          await finishForging()
        } else if (res.status === 'failed') {
          settled = true
          clearAll()
          setPhase('failed')
          setErrorMsg(res.error ?? 'Scoring failed — check your GitHub connection.')
        }
        // 'running' and 'idle' → keep polling
      } catch {
        // transient — keep polling
      }
    }

    poll()
    pollRef.current = setInterval(poll, 2500)

    // Hard timeout: 7 minutes (scoring should never take this long)
    const timeout = setTimeout(() => {
      if (!settled) {
        settled = true
        clearAll()
        setPhase('failed')
        setErrorMsg('Scoring is taking longer than expected. Please try again from your hub.')
      }
    }, 7 * 60 * 1000)
    stepTimers.current.push(timeout)

    return clearAll
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (phase === 'failed') {
    return (
      <div className="min-h-screen bg-void-950 flex flex-col items-center justify-center px-6 gap-8">
        <p
          className="text-gold-400 text-sm tracking-[0.4em] uppercase"
          style={{ fontFamily: '"Cinzel", serif' }}
        >
          Forging interrupted
        </p>
        <p className="text-ink-400 text-sm text-center max-w-sm">{errorMsg}</p>
        <button
          onClick={() => navigate('/hub', { replace: true })}
          className="text-xs tracking-widest uppercase px-6 py-2.5 transition-colors hover:text-gold-300"
          style={{ border: '1px solid rgba(100,80,30,0.5)', color: '#8a7e6a' }}
        >
          Return to Hub
        </button>
      </div>
    )
  }

  const isDone = phase === 'done'
  const isGenerating = phase === 'generating'

  return (
    <div
      className="min-h-screen flex flex-col items-center justify-center px-6 relative overflow-hidden"
      style={{ background: '#0e0a05' }}
    >
      {/* Ambient glow */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          background: 'radial-gradient(ellipse 60% 50% at 50% 55%, rgba(180,130,30,0.09) 0%, transparent 70%)',
        }}
      />

      <div className="relative z-10 flex flex-col items-center gap-10 max-w-sm w-full">
        {/* Sigil */}
        <div style={{ opacity: isDone ? 0 : 1, transition: 'opacity 0.8s' }}>
          <ForgeSigil />
        </div>

        {/* Headline */}
        <div className="text-center space-y-2">
          <h1
            className="text-2xl tracking-[0.15em]"
            style={{ fontFamily: '"Cinzel", serif', color: '#e8c040' }}
          >
            {isDone
              ? 'Your build is ready'
              : isGenerating
              ? 'Consulting the Oracle…'
              : 'Forging your character'}
          </h1>
          {!isDone && (
            <p className="text-xs tracking-widest" style={{ color: '#4a4032' }}>
              This may take a minute
            </p>
          )}
        </div>

        {/* Step list */}
        {!isDone && (
          <div className="w-full space-y-3">
            {COSMETIC_STEPS.map((step, i) => {
              const visible = i < visibleSteps
              const complete = i < visibleSteps - 1 || isGenerating
              if (!visible) return null
              return (
                <div
                  key={i}
                  className="flex items-center gap-3 animate-step-appear"
                  style={{ animationDelay: '0ms' }}
                >
                  <span style={{ color: complete ? '#5ab870' : '#c99224', fontSize: 11, flexShrink: 0 }}>
                    {complete ? '✓' : '◈'}
                  </span>
                  <span
                    className="text-xs tracking-wide"
                    style={{ color: complete ? '#8a7e6a' : '#c8bca8' }}
                  >
                    {step}
                  </span>
                </div>
              )
            })}
          </div>
        )}

        {/* Separator line at bottom */}
        <div className="w-full" style={{ height: 1, background: 'rgba(160,112,24,0.15)' }} />

        <p
          className="text-[9px] tracking-[0.3em] uppercase text-center"
          style={{ color: 'rgba(74,64,50,0.7)', fontFamily: '"Cinzel", serif' }}
        >
          ◈ Signal
        </p>
      </div>
    </div>
  )
}
