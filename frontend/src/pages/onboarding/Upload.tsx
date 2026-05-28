import { useRef, useState, useEffect, useCallback, type DragEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { uploadCV, getCVStatus } from '../../api/onboarding'
import { Button } from '../../components/ui/Button'
import type { CVData } from '../../types'

// ─── Types ────────────────────────────────────────────────────────────────────

type Phase =
  | { kind: 'idle' }
  | { kind: 'selected'; file: File }
  | { kind: 'uploading' }
  | { kind: 'processing' }
  | { kind: 'done'; data: CVData }
  | { kind: 'error'; message: string }

// ─── Scan messages cycled during AI processing ────────────────────────────────

const SCAN_MESSAGES = [
  'Reading your professional history...',
  'Identifying skill patterns...',
  'Mapping your experience depth...',
  'Inferring specializations...',
  'Analyzing contribution signals...',
  'Assembling your profile data...',
]

// ─── Component ────────────────────────────────────────────────────────────────

export function Upload() {
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)
  const [phase, setPhase] = useState<Phase>({ kind: 'idle' })
  const [dragOver, setDragOver] = useState(false)
  const [scanIdx, setScanIdx] = useState(0)

  // Cycle scan messages while processing
  useEffect(() => {
    if (phase.kind !== 'processing') return
    const id = setInterval(() => setScanIdx(i => (i + 1) % SCAN_MESSAGES.length), 2200)
    return () => clearInterval(id)
  }, [phase.kind])

  // Poll backend for CV parse status
  useEffect(() => {
    if (phase.kind !== 'processing') return
    const id = setInterval(async () => {
      try {
        const status = await getCVStatus()
        if (status.status === 'done' && status.extracted_data) {
          setPhase({ kind: 'done', data: status.extracted_data })
        } else if (status.status === 'failed') {
          setPhase({ kind: 'error', message: status.error_message ?? 'Analysis failed. Try again.' })
        }
      } catch {
        // network hiccup — keep polling
      }
    }, 2000)
    return () => clearInterval(id)
  }, [phase.kind])

  const handleFile = useCallback((file: File) => {
    if (!file.name.toLowerCase().endsWith('.pdf')) {
      setPhase({ kind: 'error', message: 'Only PDF files are supported.' })
      return
    }
    setPhase({ kind: 'selected', file })
  }, [])

  const onDrop = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) handleFile(file)
  }, [handleFile])

  async function submit() {
    if (phase.kind !== 'selected') return
    setPhase({ kind: 'uploading' })
    try {
      await uploadCV(phase.file)
      setPhase({ kind: 'processing' })
    } catch (err) {
      setPhase({ kind: 'error', message: err instanceof Error ? err.message : 'Upload failed.' })
    }
  }

  return (
    <div className="min-h-screen bg-void-950 flex flex-col items-center justify-center px-6 py-16">

      {/* Step indicator */}
      <div className="w-full max-w-lg mb-12">
        <div className="flex items-center gap-3">
          {['Upload CV', 'Connect GitHub', 'Build Reveal'].map((label, i) => (
            <div key={label} className="flex items-center gap-3 flex-1 last:flex-none">
              <div className={`flex items-center gap-2 ${i === 0 ? 'text-gold-400' : 'text-ink-600'}`}>
                <span className="text-xs font-mono border border-current w-6 h-6 flex items-center justify-center">
                  {String(i + 1).padStart(2, '0')}
                </span>
                <span className="text-xs tracking-wide hidden sm:block">{label}</span>
              </div>
              {i < 2 && <div className="flex-1 h-px bg-void-700" />}
            </div>
          ))}
        </div>
      </div>

      {/* Card */}
      <div className="w-full max-w-lg">
        <div className="mb-8">
          <p className="text-gold-400 text-xs tracking-[0.3em] uppercase mb-3">Step 01</p>
          <h1 className="text-2xl font-bold text-ink-50 tracking-tight">Upload your CV</h1>
          <p className="mt-2 text-sm text-ink-400">
            We extract your experience and skills — no manual entry.
          </p>
        </div>

        {/* ── Idle / Selected ── */}
        {(phase.kind === 'idle' || phase.kind === 'selected') && (
          <div className="flex flex-col gap-6">
            <div
              role="button"
              tabIndex={0}
              onClick={() => inputRef.current?.click()}
              onKeyDown={e => e.key === 'Enter' && inputRef.current?.click()}
              onDragOver={e => { e.preventDefault(); setDragOver(true) }}
              onDragLeave={() => setDragOver(false)}
              onDrop={onDrop}
              className={`
                border-2 border-dashed p-12 text-center cursor-pointer transition-all
                ${dragOver
                  ? 'border-gold-400 bg-gold-400/5'
                  : phase.kind === 'selected'
                    ? 'border-gold-500/60 bg-void-900'
                    : 'border-void-600 hover:border-void-500 bg-void-900'
                }
              `}
            >
              {phase.kind === 'selected' ? (
                <div className="flex flex-col items-center gap-3">
                  <span className="text-gold-400 text-2xl">◈</span>
                  <span className="text-sm font-medium text-ink-50">{phase.file.name}</span>
                  <span className="text-xs text-ink-400">
                    {(phase.file.size / 1024).toFixed(0)} KB — click to replace
                  </span>
                </div>
              ) : (
                <div className="flex flex-col items-center gap-3">
                  <span className="text-void-500 text-3xl">↑</span>
                  <span className="text-sm text-ink-400">
                    Drop your PDF here, or{' '}
                    <span className="text-gold-400 underline underline-offset-2">browse</span>
                  </span>
                  <span className="text-xs text-ink-600">PDF only · Max 10 MB</span>
                </div>
              )}
            </div>

            <input
              ref={inputRef}
              type="file"
              accept=".pdf"
              className="hidden"
              onChange={e => e.target.files?.[0] && handleFile(e.target.files[0])}
            />

            <Button
              onClick={submit}
              disabled={phase.kind !== 'selected'}
              className="w-full py-3"
            >
              Analyze CV →
            </Button>
          </div>
        )}

        {/* ── Uploading ── */}
        {phase.kind === 'uploading' && (
          <div className="border border-void-700 bg-void-900 p-12 text-center flex flex-col items-center gap-4">
            <div className="w-6 h-6 rounded-full border-2 border-gold-400 border-t-transparent animate-spin" />
            <span className="text-sm text-ink-400">Uploading...</span>
          </div>
        )}

        {/* ── Processing (AI scanning) ── */}
        {phase.kind === 'processing' && (
          <div className="border border-void-700 bg-void-900 p-12 flex flex-col items-center gap-6">
            {/* Animated scan grid */}
            <div className="relative w-16 h-16">
              <div className="absolute inset-0 border border-gold-400/30 animate-ping" />
              <div className="absolute inset-2 border border-gold-400/50" />
              <div className="absolute inset-0 flex items-center justify-center text-gold-400 text-xl">◈</div>
            </div>
            <div className="text-center">
              <p className="text-sm text-ink-200 mb-1 transition-all">{SCAN_MESSAGES[scanIdx]}</p>
              <p className="text-xs text-ink-600">This takes 10–20 seconds</p>
            </div>
            {/* Scanning bar */}
            <div className="w-full h-px bg-void-700 relative overflow-hidden">
              <div className="absolute inset-y-0 left-0 w-1/3 bg-gold-400/40 animate-[scan_2s_ease-in-out_infinite]" />
            </div>
          </div>
        )}

        {/* ── Done ── */}
        {phase.kind === 'done' && (
          <div className="flex flex-col gap-4">
            <div className="border border-void-700 bg-void-900 p-8 flex flex-col gap-6">
              <div className="flex items-center gap-3">
                <span className="text-signal-400 text-lg">◈</span>
                <span className="text-sm font-medium text-ink-50">Analysis complete</span>
              </div>

              {phase.data.full_name && (
                <p className="text-xl font-bold text-ink-50">{phase.data.full_name}</p>
              )}

              <div className="grid grid-cols-2 gap-3">
                <Stat label="Experiences" value={phase.data.experiences.length} />
                <Stat label="Skills detected" value={phase.data.skills.length} />
              </div>

              {phase.data.inferred_specializations.length > 0 && (
                <div>
                  <p className="text-xs text-ink-600 tracking-widest uppercase mb-3">
                    Inferred specializations
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {phase.data.inferred_specializations.map(s => (
                      <span
                        key={s}
                        className="text-xs border border-gold-500/40 text-gold-400 px-3 py-1 tracking-wide"
                      >
                        {s}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <Button onClick={() => navigate('/onboarding/reveal')} className="w-full py-3">
              Continue to build reveal →
            </Button>

            <p className="text-xs text-ink-600 text-center">
              You can review and correct everything before publishing.
            </p>
          </div>
        )}

        {/* ── Error ── */}
        {phase.kind === 'error' && (
          <div className="flex flex-col gap-4">
            <div className="border border-danger-400/40 bg-void-900 p-8">
              <p className="text-sm text-danger-400 mb-1">Something went wrong</p>
              <p className="text-sm text-ink-400">{phase.message}</p>
            </div>
            <Button variant="ghost" onClick={() => setPhase({ kind: 'idle' })} className="w-full">
              Try again
            </Button>
          </div>
        )}
      </div>

      {/* Skip link */}
      {(phase.kind === 'idle' || phase.kind === 'selected') && (
        <button
          onClick={() => navigate('/hub')}
          className="mt-10 text-xs text-ink-600 hover:text-ink-400 transition-colors"
        >
          Skip for now →
        </button>
      )}
    </div>
  )
}

// ─── Small stat block ─────────────────────────────────────────────────────────

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="border border-void-700 p-4">
      <p className="text-2xl font-bold text-gold-400">{value}</p>
      <p className="text-xs text-ink-400 mt-1">{label}</p>
    </div>
  )
}
