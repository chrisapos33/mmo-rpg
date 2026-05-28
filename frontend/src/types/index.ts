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
