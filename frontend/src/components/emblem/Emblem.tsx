import type { UserSignalScore } from '../../types'

// ─── Seeded PRNG ──────────────────────────────────────────────────────────────

function fnvHash(s: string): number {
  let h = 2166136261 >>> 0
  for (let i = 0; i < s.length; i++) {
    h = Math.imul(h ^ s.charCodeAt(i), 16777619) >>> 0
  }
  return h
}

function mulberry32(seed: number) {
  let s = seed >>> 0
  return (): number => {
    s = (s + 0x6D2B79F5) >>> 0
    let t = Math.imul(s ^ (s >>> 15), 1 | s)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

// ─── Dimension config ─────────────────────────────────────────────────────────

// Pentagon of petals, top going clockwise. Each dimension has a fixed angle and color.
const DIMS = [
  { key: 'output_percentile'        as const, angle: -90, color: '#c99224', glow: '#f0d060' },
  { key: 'craft_percentile'         as const, angle: -18, color: '#5a90d4', glow: '#8ab8f0' },
  { key: 'influence_percentile'     as const, angle:  54, color: '#f0d040', glow: '#fff0a0' },
  { key: 'collaboration_percentile' as const, angle: 126, color: '#5ab870', glow: '#90d8a0' },
  { key: 'range_percentile'         as const, angle: 198, color: '#c45a70', glow: '#e890a0' },
]

// ─── Geometry helpers ─────────────────────────────────────────────────────────

function octPts(cx: number, cy: number, r: number): string {
  return Array.from({ length: 8 }, (_, i) => {
    const a = (i * 45 - 22.5) * (Math.PI / 180)
    return `${(cx + r * Math.cos(a)).toFixed(2)},${(cy + r * Math.sin(a)).toFixed(2)}`
  }).join(' ')
}

function polar(cx: number, cy: number, r: number, angleDeg: number) {
  const a = angleDeg * (Math.PI / 180)
  return { x: cx + r * Math.cos(a), y: cy + r * Math.sin(a) }
}

// ─── Class icon paths (24×24 viewbox, stroke-only) ───────────────────────────

const CLASS_PATHS: Record<string, string> = {
  'The Architect':  'M2 22h20M5 22V10L12 3l7 7v12M10 22v-6h4v6',
  'The Artisan':    'M9 3l1.5 1.5-7 7 1.5 1.5 7-7 1.5 1.5M15 3l6 6M3 21l6-6',
  'The Pathfinder': 'M12 2l2.5 5 5.5.8-4 3.9.9 5.4L12 14.8l-4.9 2.3.9-5.4-4-3.9 5.5-.8z',
  'The Sage':       'M4 4v16h16V8l-4-4H4zm8 0v4h4M8 12h8M8 16h5',
  'The Operator':   'M4 4h16v16H4zM8 9l4 3-4 3M14 15h4',
  'The Sentinel':   'M12 2L4 6v6c0 5 3.5 9.7 8 11 4.5-1.3 8-6 8-11V6l-8-4zM9 12l2 2 4-4',
  'The Artificer':  'M12 6a6 6 0 100 12 6 6 0 000-12zm0 2a4 4 0 110 8 4 4 0 010-8zM12 2v2M12 20v2M2 12h2M20 12h2M5.64 5.64l1.42 1.42M17 17l1.42 1.42M5.64 18.36l1.42-1.42M17 7l1.42-1.42',
}

// ─── Emblem ───────────────────────────────────────────────────────────────────

interface Props {
  score: UserSignalScore
  cls: string
  size?: number
}

export function EmblemSVG({ score, cls, size = 200 }: Props) {
  const uid = score.user_id

  // Deterministic seed from stable identifiers
  const seed = fnvHash(
    [uid, score.output_percentile, score.craft_percentile, score.influence_percentile,
     score.collaboration_percentile, score.range_percentile].map(v => Math.round(+v)).join(':')
  )
  const rand = mulberry32(seed)

  const CX = 100, CY = 100
  const OUTER_R = 92
  const INNER_R  = 76

  // Dominant dimension drives the radial background color
  const dimVals = DIMS.map(d => ({ ...d, p: score[d.key] }))
  const dominant = [...dimVals].sort((a, b) => b.p - a.p)[0]

  // Trust gems: 0–5, one filled per ~20% trust
  const trustTier = Math.min(5, Math.round((score.trust ?? 0) * 5))

  // Petal half-lengths
  const MIN_H = 14, MAX_H = 54
  const PETAL_RX = 9

  // Procedural corner rune type (0=cross, 1=diamond, 2=circle) — seeded per corner
  const cornerAngles = [0, 90, 180, 270]
  const cornerTypes = cornerAngles.map(() => Math.floor(rand() * 3))

  // Additional inner accent rings (1–2 concentric thin rings, radius varies per user)
  const ringR = INNER_R * 0.58 + rand() * INNER_R * 0.12

  const bgGradId  = `bg-${uid}`
  const clipId    = `clip-${uid}`
  const iconPath  = CLASS_PATHS[cls] ?? CLASS_PATHS['The Pathfinder']

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 200 200"
      style={{ display: 'block', overflow: 'visible' }}
      aria-label={`${cls} emblem`}
    >
      <defs>
        {/* Octagonal clip */}
        <clipPath id={clipId}>
          <polygon points={octPts(CX, CY, OUTER_R)} />
        </clipPath>

        {/* Radial glow based on dominant dimension */}
        <radialGradient id={bgGradId} cx="50%" cy="50%" r="50%">
          <stop offset="0%"   stopColor={dominant.color} stopOpacity="0.18" />
          <stop offset="60%"  stopColor="#0e0a05"         stopOpacity="0.85" />
          <stop offset="100%" stopColor="#0a0703"         stopOpacity="1" />
        </radialGradient>
      </defs>

      {/* ── Background field ──────────────────────────────────────────── */}
      <polygon points={octPts(CX, CY, OUTER_R)} fill="#0f0b06" />
      <polygon points={octPts(CX, CY, OUTER_R)} fill={`url(#${bgGradId})`} />

      {/* Subtle woven grid texture */}
      <g clipPath={`url(#${clipId})`} opacity="0.055">
        {Array.from({ length: 12 }, (_, i) => (
          <line key={`h${i}`} x1="8" y1={8 + i * 16} x2="192" y2={8 + i * 16} stroke="#c99224" strokeWidth="0.5" />
        ))}
        {Array.from({ length: 12 }, (_, i) => (
          <line key={`v${i}`} x1={8 + i * 16} y1="8" x2={8 + i * 16} y2="192" stroke="#c99224" strokeWidth="0.5" />
        ))}
      </g>

      {/* ── Dimension petals ──────────────────────────────────────────── */}
      <g clipPath={`url(#${clipId})`}>
        {dimVals.map(d => {
          const halfLen = MIN_H + (d.p / 100) * (MAX_H - MIN_H)
          const petalCY = CY - halfLen
          const opacity = 0.38 + 0.52 * (d.p / 100)
          return (
            <g key={d.key} transform={`rotate(${d.angle}, ${CX}, ${CY})`}>
              {/* Main petal body */}
              <ellipse cx={CX} cy={petalCY} rx={PETAL_RX} ry={halfLen} fill={d.color} opacity={opacity} />
              {/* Inner highlight streak */}
              <ellipse
                cx={CX}
                cy={petalCY + halfLen * 0.38}
                rx={3.5}
                ry={halfLen * 0.38}
                fill={d.glow}
                opacity={0.18 + 0.15 * (d.p / 100)}
              />
            </g>
          )
        })}
      </g>

      {/* ── Inner accent ring (seeded radius) ────────────────────────── */}
      <circle
        cx={CX} cy={CY} r={ringR}
        fill="none"
        stroke="rgba(160,112,24,0.14)"
        strokeWidth="0.6"
      />

      {/* ── Central medallion ─────────────────────────────────────────── */}
      <circle cx={CX} cy={CY} r={31} fill="#0f0b06" />
      <circle
        cx={CX} cy={CY} r={29}
        fill="rgba(10,7,2,0.9)"
        stroke={dominant.color}
        strokeWidth="1.3"
        strokeOpacity="0.85"
      />

      {/* Class icon — stroked, gold, centered in medallion */}
      <g transform={`translate(${CX - 14}, ${CY - 14}) scale(${28 / 24})`}>
        <path
          d={iconPath}
          stroke="#e8c040"
          strokeWidth="1.4"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
          strokeOpacity="0.92"
        />
      </g>

      {/* ── Structural rings ──────────────────────────────────────────── */}
      {/* Inner ring */}
      <polygon
        points={octPts(CX, CY, INNER_R)}
        fill="none"
        stroke="rgba(160,112,24,0.28)"
        strokeWidth="0.7"
      />
      {/* Outer ring */}
      <polygon
        points={octPts(CX, CY, OUTER_R)}
        fill="none"
        stroke="#c99224"
        strokeWidth="1.5"
      />

      {/* 8 tick marks at octagon corners */}
      {Array.from({ length: 8 }, (_, i) => {
        const inner = polar(CX, CY, OUTER_R - 8, i * 45 - 22.5)
        const outer = polar(CX, CY, OUTER_R,     i * 45 - 22.5)
        return (
          <line key={i} x1={inner.x} y1={inner.y} x2={outer.x} y2={outer.y}
            stroke="#c99224" strokeWidth="1.8" />
        )
      })}

      {/* ── Corner rune decorations ────────────────────────────────────── */}
      {cornerAngles.map((a, i) => {
        const { x: rx, y: ry } = polar(CX, CY, INNER_R - 13, a)
        const s = 5
        const t = cornerTypes[i]
        if (t === 0) {
          return (
            <g key={i} opacity="0.42">
              <line x1={rx - s} y1={ry} x2={rx + s} y2={ry} stroke="#c99224" strokeWidth="1" />
              <line x1={rx} y1={ry - s} x2={rx} y2={ry + s} stroke="#c99224" strokeWidth="1" />
            </g>
          )
        }
        if (t === 1) {
          return (
            <polygon key={i}
              points={`${rx},${ry - s} ${rx + s},${ry} ${rx},${ry + s} ${rx - s},${ry}`}
              fill="none" stroke="#c99224" strokeWidth="0.8" opacity="0.4" />
          )
        }
        return (
          <circle key={i} cx={rx} cy={ry} r={s * 0.65}
            fill="none" stroke="#c99224" strokeWidth="0.8" opacity="0.38" />
        )
      })}

      {/* ── Trust gem indicators (bottom arc, 5 positions) ────────────── */}
      {Array.from({ length: 5 }, (_, i) => {
        const angleDeg = 108 + i * 36
        const { x: gx, y: gy } = polar(CX, CY, OUTER_R - 10, angleDeg)
        const filled = i < trustTier
        const s = 4.5
        return (
          <polygon key={i}
            points={`${gx},${gy - s} ${gx + s},${gy} ${gx},${gy + s} ${gx - s},${gy}`}
            fill={filled ? '#c99224' : 'none'}
            stroke="#c99224"
            strokeWidth={filled ? 0 : 0.8}
            opacity={filled ? 0.92 : 0.22}
          />
        )
      })}
    </svg>
  )
}
