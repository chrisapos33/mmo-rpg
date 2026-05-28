import { api } from './client'
import type { ExploreResponse } from '../types'

export interface ExploreFilters {
  class?: string
  sort?: 'signal' | 'recent'
  limit?: number
  offset?: number
}

export function getExplore(filters: ExploreFilters = {}): Promise<ExploreResponse> {
  const params = new URLSearchParams()
  if (filters.class)  params.set('class',  filters.class)
  if (filters.sort)   params.set('sort',   filters.sort)
  if (filters.limit)  params.set('limit',  String(filters.limit))
  if (filters.offset) params.set('offset', String(filters.offset))

  const qs = params.toString()
  return api.get<ExploreResponse>(`/explore${qs ? `?${qs}` : ''}`)
}
