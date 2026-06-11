export interface User {
  id: number
  email: string
  created_at: string
  updated_at: string
}

export interface Profile {
  id: number
  user_id: number
  username: string | null
  display_name: string | null
  class: string | null
  subclass: string | null
  headline: string | null
  summary: string | null
  avatar_url: string | null
  signal_score: number
  xp: number
  is_published: boolean
  onboarding_step: string
  strengths: string[]
  growth_paths: string[]
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiError {
  error: string
}

// ─── CV / Onboarding ─────────────────────────────────────────────────────────

export interface CVExperience {
  company: string
  title: string
  start_date: string
  end_date: string | null
  is_current: boolean
  description: string | null
}

export interface CVEducation {
  institution: string
  degree: string | null
  field: string | null
  year: string | null
}

export interface CVData {
  full_name: string
  email: string | null
  location: string | null
  summary: string | null
  experiences: CVExperience[]
  skills: string[]
  education: CVEducation[]
  languages: string[]
  inferred_specializations: string[]
}

export interface CVUploadStatus {
  id: number
  status: 'processing' | 'done' | 'failed'
  original_name: string
  extracted_data: CVData | null
  error_message: string | null
  created_at: string
  processed_at: string | null
}

// ─── GitHub ──────────────────────────────────────────────────────────────────

export interface GitHubConnection {
  id: number
  user_id: number
  github_username: string
  github_user_id: number
  avatar_url: string | null
  repo_count: number
  star_count: number
  followers: number
  top_languages: string[]
  contribution_score: number
  synced_at: string | null
  created_at: string
  updated_at: string
}

export interface GitHubStatusResponse {
  connected: boolean
  connection: GitHubConnection | null
}

// ─── Signal ──────────────────────────────────────────────────────────────────

export type SignalDimension = 'output' | 'craft' | 'influence' | 'collaboration' | 'range'

export interface UserSignalScore {
  user_id: number
  github_username?: string | null
  output_percentile: number
  craft_percentile: number
  influence_percentile: number
  collaboration_percentile: number
  range_percentile: number
  trust: number
  scoring_status?: string | null
  scoring_started_at?: string | null
  scoring_done_at?: string | null
  scoring_error?: string | null
  updated_at: string
}

// ─── Explore ─────────────────────────────────────────────────────────────────

export interface ExploreEntry {
  user_id: number
  class: string | null
  subclass: string | null
  headline: string | null
  trust: number
  github_username: string | null
  top_languages: string[]
  updated_at: string
}

export interface ExploreResponse {
  entries: ExploreEntry[]
  classes: string[]
  limit: number
  offset: number
}

export interface PublicProfileResponse {
  profile: Profile
  signal: UserSignalScore | null
  github: GitHubConnection | null
}

export type EvidenceSourceType = 'blog' | 'portfolio' | 'community' | 'other'

export const EVIDENCE_SOURCE_LABELS: Record<EvidenceSourceType, string> = {
  blog:      'Blog / Writing',
  portfolio: 'Portfolio Project',
  community: 'Community Contribution',
  other:     'Other',
}

export interface EvidenceItem {
  id: number
  user_id: number
  source_type: EvidenceSourceType
  source_key: string
  artifact_url: string | null
  title: string
  description: string | null
  metadata: Record<string, unknown> | null
  verification_status: 'unverified' | 'url_verified' | 'platform_verified' | 'peer_verified' | 'admin_verified'
  verification_confidence: number
  created_at: string
  updated_at: string
}

export interface SubmitEvidenceRequest {
  source_type: EvidenceSourceType
  title: string
  artifact_url: string
  description?: string
}
