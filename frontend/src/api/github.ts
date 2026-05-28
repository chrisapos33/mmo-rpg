import { api } from './client'
import type { GitHubStatusResponse, GitHubConnection } from '../types'

export async function getGitHubAuthorizeUrl(): Promise<string> {
  const res = await api.get<{ url: string }>('/github/authorize')
  return res.url
}

export async function getGitHubStatus(): Promise<GitHubStatusResponse> {
  return api.get<GitHubStatusResponse>('/github/status')
}

export async function syncGitHub(): Promise<GitHubConnection> {
  return api.post<GitHubConnection>('/github/sync', {})
}
