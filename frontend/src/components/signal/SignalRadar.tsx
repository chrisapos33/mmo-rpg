import type { UserSignalScore, SignalDimension } from '../../types'

// Dimensions in clockwise order from top.
const DIMENSIONS: { key: SignalDimension; label: string; description: string }[] = [
  { key: 'output',        label: 'Output',    description: 'Active shipping on validated repos' },
  { key: 'craft',         label: 'Craft',     description: 'Tests, CI, review depth, longevity' },
  { key: 'influence',     label: 'Influence', description: 'Dependents, stars, forks from others' },
  { key: 'collaboration', label: 'Collab',    description: 'External merged PRs and reviews' },
  { key: 'range',         label: 'Range',     description: 'Validated breadth across stacks' },
]

const N = DIMENSIONS.length
const CX = 140
const CY = 140
const R = 100
const LEVELS = 5

function polarToCart(angle: number, radius: number): [number, number] {
  // angle=0 is top (−π/2 offset), clockwise
  const rad = (angle - 90) * (Math.PI / 180)
  return [CX + radius * Math.cos(rad), CY + radius * Math.sin(rad)]
}

function axisAngle(i: number) {
  return (360 / N) * i
}

function scoreForDim(scores: UserSignalScore, dim: SignalDimension): number {
  const key = `${dim}_percentile` as keyof UserSignalScore
  return (scores[key] as number) ?? 0
}

interface Props {
  scores: UserSignalScore
  size?: number
}

export function SignalRadar({ scores }: Props) {
  // Build grid hexagons (5 rings)
  const gridPolygons = Array.from({ length: LEVELS }, (_, level) => {
    const r = R * ((level + 1) / LEVELS)
    const pts = DIMENSIONS.map((_, i) => polarToCart(axisAngle(i), r))
    return pts.map(([x, y]) => `${x},${y}`).join(' ')
  })

  // Build score polygon
  const scorePoints = DIMENSIONS.map(({ key }, i) => {
    const score = Math.min(100, Math.max(0, scoreForDim(scores, key)))
    const r = R * (score / 100)
    return polarToCart(axisAngle(i), r)
  })
  const scorePath = scorePoints.map(([x, y]) => `${x},${y}`).join(' ')

  // Label positions — push slightly beyond the axis tip
  const labelPositions = DIMENSIONS.map(({ label, description }, i) => {
    const angle = axisAngle(i)
    const [x, y] = polarToCart(angle, R + 22)
    const anchor: 'start' | 'end' | 'middle' = angle < 10 || angle > 350 ? 'middle'
      : angle < 190 ? 'start'
      : angle > 190 ? 'end'
      : 'middle'
    return { x, y, label, description, anchor }
  })

  const trustPct = Math.round((scores.trust ?? 0) * 100)

  return (
    <div className="flex flex-col lg:flex-row items-center gap-8 lg:gap-12">
      {/* SVG radar */}
      <div className="flex-shrink-0">
        <svg
          viewBox="0 0 280 280"
          width={280}
          height={280}
          aria-label="Signal dimension radar"
        >
          {/* Grid rings */}
          {gridPolygons.map((pts, level) => (
            <polygon
              key={level}
              points={pts}
              fill="none"
              stroke={level === LEVELS - 1 ? '#3d3d52' : '#2a2a3a'}
              strokeWidth={level === LEVELS - 1 ? 1 : 0.5}
            />
          ))}

          {/* Axis lines */}
          {DIMENSIONS.map((_, i) => {
            const [x, y] = polarToCart(axisAngle(i), R)
            return (
              <line
                key={i}
                x1={CX} y1={CY}
                x2={x} y2={y}
                stroke="#2a2a3a"
                strokeWidth={0.75}
              />
            )
          })}

          {/* Score polygon fill */}
          <polygon
            points={scorePath}
            fill="rgba(180,140,60,0.12)"
            stroke="#b48c3c"
            strokeWidth={1.5}
            strokeLinejoin="round"
          />

          {/* Score dots */}
          {scorePoints.map(([x, y], i) => {
            const score = scoreForDim(scores, DIMENSIONS[i].key)
            if (score === 0) return null
            return (
              <circle key={i} cx={x} cy={y} r={3} fill="#d4a853" />
            )
          })}

          {/* Center dot */}
          <circle cx={CX} cy={CY} r={2} fill="#3d3d52" />

          {/* Axis labels */}
          {labelPositions.map(({ x, y, label, anchor }) => (
            <text
              key={label}
              x={x} y={y}
              textAnchor={anchor}
              dominantBaseline="middle"
              fontSize={9}
              fontFamily="ui-monospace, monospace"
              letterSpacing="0.08em"
              fill="#9898b0"
              style={{ textTransform: 'uppercase' }}
            >
              {label}
            </text>
          ))}
        </svg>
      </div>

      {/* Dimension breakdown list */}
      <div className="flex-1 w-full space-y-3">
        <div className="flex items-baseline justify-between mb-5">
          <span className="text-xs text-gold-400 tracking-[0.25em] uppercase">Trust</span>
          <span className="text-3xl font-bold text-ink-50 tabular-nums">{trustPct}%</span>
        </div>

        {DIMENSIONS.map(({ key, label, description }) => {
          const score = scoreForDim(scores, key)
          const pct = Math.min(100, score)
          const hasSignal = score > 0
          return (
            <div key={key} className="space-y-1">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-medium text-ink-300 tracking-wide uppercase w-24">
                    {label}
                  </span>
                  {!hasSignal && (
                    <span className="text-xs text-ink-600">no signal yet</span>
                  )}
                </div>
                <span className="text-xs font-mono text-ink-400 tabular-nums">{score}</span>
              </div>
              <div className="h-px w-full bg-void-700">
                {hasSignal && (
                  <div
                    className="h-full bg-gold-400/60 transition-all duration-700"
                    style={{ width: `${pct}%` }}
                  />
                )}
              </div>
              {!hasSignal && (
                <p className="text-xs text-ink-600 leading-snug">{description}</p>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
