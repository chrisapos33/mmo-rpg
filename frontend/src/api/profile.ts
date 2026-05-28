import { api } from './client'
import type { PublicProfileResponse } from '../types'

export function publishProfile(): Promise<{ published: boolean; profile_url: string }> {
  return api.post('/profile/publish', {})
}

export function getPublicProfile(userID: string | number): Promise<PublicProfileResponse> {
  return api.get<PublicProfileResponse>(`/p/${userID}`)
}
