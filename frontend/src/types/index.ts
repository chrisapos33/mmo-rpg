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
