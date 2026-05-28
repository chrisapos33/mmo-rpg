import { api } from './client'
import type { UserSignalScore, EvidenceItem } from '../types'

export function getSignalScores(): Promise<UserSignalScore> {
  return api.get<UserSignalScore>('/signal/scores')
}

export function getSignalEvidence(): Promise<EvidenceItem[]> {
  return api.get<EvidenceItem[]>('/signal/evidence')
}
