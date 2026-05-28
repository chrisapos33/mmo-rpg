import { api } from './client'
import type { EvidenceItem, SubmitEvidenceRequest } from '../types'

export function submitEvidence(req: SubmitEvidenceRequest): Promise<EvidenceItem> {
  return api.post<EvidenceItem>('/evidence/', req)
}

export function removeEvidence(id: number): Promise<void> {
  return api.delete<void>(`/evidence/${id}`)
}
