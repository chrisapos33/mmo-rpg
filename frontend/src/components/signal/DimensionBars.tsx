import type { ReactNode } from 'react'
import type { UserSignalScore } from '../../types'

// ─── Inline SVG icons ─────────────────────────────────────────────────────────

function SwordIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <line x1="3" y1="17" x2="14" y2="3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <line x1="14" y1="3" x2="17" y2="6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="8" y1="12" x2="6" y2="10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <circle cx="3.5" cy="16.5" r="1.5" fill="currentColor" />
    </svg>
  )
}

function HammerIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <rect x="10" y="2" width="6" height="4" rx="1" fill="currentColor" transform="rotate(45 13 4)" />
      <line x1="10" y1="8" x2="4" y2="17" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" />
    </svg>
  )
}

function CrownIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <polyline points="2,14 5,7 10,12 15,7 18,14" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round" strokeLinecap="round" fill="none" />
      <line x1="2" y1="14" x2="18" y2="14" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
    </svg>
  )
}

function ChainIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <circle cx="7" cy="10" r="4" stroke="currentColor" strokeWidth="1.5" fill="none" />
      <circle cx="13" cy="10" r="4" stroke="currentColor" strokeWidth="1.5" fill="none" />
    </svg>
  )
}

function CompassIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <circle cx="10" cy="10" r="7.5" stroke="currentColor" strokeWidth="1.5" fill="none" />
      <line x1="10" y1="3" x2="10" y2="5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="10" y1="15" x2="10" y2="17" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="3" y1="10" x2="5" y2="10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="15" y1="10" x2="17" y2="10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
      <circle cx="10" cy="10" r="1.5" fill="currentColor" />
    </svg>
  )
}

function ShieldIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <path d="M10 2L3 5v5c0 4 3 7 7 8 4-1 7-4 7-8V5L10 2Z" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinejoin="round" />
      <polyline points="7,10 9,12 13,8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

// ─── Bar fill gradients ────────────────────────────────────────────────────────

type BarColor = 'gold' | 'bronze' | 'azure' | 'forest' | 'burgundy'

const GRADIENTS: Record<BarColor, string> = {
  gold:     'linear-gradient(to bottom, #e8c040 0%, #a07018 50%, #c99224 100%)',
  bronze:   'linear-gradient(to bottom, #d4a855 0%, #7a4e18 50%, #b48638 100%)',
  azure:    'linear-gradient(to bottom, #6aa0e4 0%, #1a3a6a 50%, #5a90d4 100%)',
  forest:   'linear-gradient(to bottom, #6ac87a 0%, #1a5228 50%, #5ab870 100%)',
  burgundy: 'linear-gradient(to bottom, #d46878 0%, #5a1020 50%, #c45a70 100%)',
}

const HIGHLIGHT_SHADOWS: Record<BarColor, string> = {
  gold:     'inset 0 1px 0 rgba(255,220,100,0.25)',
  bronze:   'inset 0 1px 0 rgba(255,200,80,0.2)',
  azure:    'inset 0 1px 0 rgba(140,200,255,0.2)',
  forest:   'inset 0 1px 0 rgba(140,255,160,0.2)',
  burgundy: 'inset 0 1px 0 rgba(255,140,160,0.2)',
}

const LABEL_COLORS: Record<BarColor, string> = {
  gold:     '#c99224',
  bronze:   '#d4a855',
  azure:    '#5a90d4',
  forest:   '#5ab870',
  burgundy: '#c45a70',
}

// ─── Dimension config ─────────────────────────────────────────────────────────

interface DimensionConfig {
  key: 'output' | 'craft' | 'influence' | 'collaboration' | 'range'
  label: string
  color: BarColor
  icon: ReactNode
  scoreKey: keyof UserSignalScore
}

const DIMENSIONS: DimensionConfig[] = [
  { key: 'output',        label: 'Output',        color: 'gold',     icon: <SwordIcon />,    scoreKey: 'output_percentile' },
  { key: 'craft',         label: 'Craft',          color: 'bronze',   icon: <HammerIcon />,   scoreKey: 'craft_percentile' },
  { key: 'influence',     label: 'Influence',      color: 'azure',    icon: <CrownIcon />,    scoreKey: 'influence_percentile' },
  { key: 'collaboration', label: 'Collaboration',  color: 'forest',   icon: <ChainIcon />,    scoreKey: 'collaboration_percentile' },
  { key: 'range',         label: 'Range',          color: 'burgundy', icon: <CompassIcon />,  scoreKey: 'range_percentile' },
]

// ─── Tier helpers ─────────────────────────────────────────────────────────────

function getTierLabel(p: number): { label: string; color: string } {
  if (p >= 90) return { label: 'Elite',  color: '#e8c040' }
  if (p >= 70) return { label: 'High',   color: '#d4a855' }
  if (p >= 30) return { label: 'Mid',    color: '#c8bca8' }
  return             { label: 'Low',    color: '#8a7e6a' }
}

// ─── Single bar ───────────────────────────────────────────────────────────────

interface BarRowProps {
  icon: ReactNode
  label: string
  color: BarColor
  percentile: number
}

function BarRow({ icon, label, color, percentile }: BarRowProps) {
  const pct = Math.round(Math.min(100, Math.max(0, percentile)))
  const tier = getTierLabel(pct)
  const labelColor = LABEL_COLORS[color]

  return (
    <div className="flex items-center gap-3">
      {/* Icon + label */}
      <div
        className="flex items-center gap-1.5 flex-shrink-0"
        style={{ width: '7.5rem', color: labelColor }}
      >
        {icon}
        <span
          className="text-[10px] uppercase tracking-[0.12em] leading-none"
          style={{ fontFamily: '"Cinzel", serif' }}
        >
          {label}
        </span>
      </div>

      {/* Bar track */}
      <div
        className="flex-1 relative"
        style={{
          height: '14px',
          background: 'rgba(18,10,4,0.9)',
          boxShadow: 'inset 0 2px 4px rgba(0,0,0,0.8), inset 0 0 0 1px rgba(80,55,10,0.4)',
        }}
      >
        {pct > 0 && (
          <div
            className="absolute top-0 left-0 h-full"
            style={{
              width: `${pct}%`,
              background: GRADIENTS[color],
              boxShadow: HIGHLIGHT_SHADOWS[color],
              transition: 'width 0.6s cubic-bezier(0.16,1,0.3,1)',
            }}
          />
        )}
      </div>

      {/* Percentile + tier */}
      <div className="flex items-baseline gap-1 flex-shrink-0" style={{ minWidth: '5.5rem', textAlign: 'right', justifyContent: 'flex-end', display: 'flex' }}>
        <span
          className="text-base font-bold tabular-nums leading-none"
          style={{ color: labelColor, fontVariantNumeric: 'tabular-nums' }}
        >
          {pct}
          <sup className="text-[8px] ml-0.5" style={{ color: labelColor }}>p</sup>
        </span>
        <span
          className="text-[9px] uppercase tracking-widest"
          style={{ color: tier.color }}
        >
          {tier.label}
        </span>
      </div>
    </div>
  )
}

// ─── Trust bar ────────────────────────────────────────────────────────────────

function TrustBar({ trust }: { trust: number }) {
  const pct = Math.min(100, Math.max(0, trust * 100))

  return (
    <div className="mt-4 pt-4" style={{ borderTop: '1px solid rgba(200,150,50,0.2)' }}>
      <div className="flex items-center gap-3">
        {/* Label */}
        <div
          className="flex items-center gap-1.5 flex-shrink-0"
          style={{ width: '7.5rem', color: '#c99224' }}
        >
          <ShieldIcon />
          <span
            className="text-[10px] uppercase tracking-[0.12em]"
            style={{ fontFamily: '"Cinzel", serif' }}
          >
            Trust
          </span>
        </div>

        {/* Bar */}
        <div
          className="flex-1 relative"
          style={{
            height: '14px',
            background: 'rgba(18,10,4,0.9)',
            boxShadow: 'inset 0 2px 4px rgba(0,0,0,0.8), inset 0 0 0 1px rgba(80,55,10,0.4)',
          }}
        >
          {pct > 0 && (
            <div
              className="absolute top-0 left-0 h-full"
              style={{
                width: `${pct}%`,
                background: GRADIENTS.gold,
                boxShadow: HIGHLIGHT_SHADOWS.gold,
                transition: 'width 0.6s cubic-bezier(0.16,1,0.3,1)',
              }}
            />
          )}
        </div>

        {/* Value */}
        <div className="flex-shrink-0" style={{ minWidth: '5.5rem', textAlign: 'right' }}>
          <span
            className="text-base font-bold tabular-nums"
            title="Fraction of build resting on verified GitHub evidence"
            style={{ color: '#c99224', cursor: 'help' }}
          >
            {pct.toFixed(1)}
            <span className="text-[10px] ml-0.5">%</span>
          </span>
        </div>
      </div>
      <p className="mt-1.5 text-[9px] tracking-wide" style={{ color: '#4a4032', marginLeft: '9rem' }}>
        Verified GitHub evidence
      </p>
    </div>
  )
}

// ─── DimensionBars ────────────────────────────────────────────────────────────

interface Props {
  scores: UserSignalScore
}

export function DimensionBars({ scores }: Props) {
  return (
    <div className="space-y-3">
      {DIMENSIONS.map(dim => (
        <BarRow
          key={dim.key}
          icon={dim.icon}
          label={dim.label}
          color={dim.color}
          percentile={scores[dim.scoreKey] as number}
        />
      ))}
      <TrustBar trust={scores.trust} />
    </div>
  )
}
