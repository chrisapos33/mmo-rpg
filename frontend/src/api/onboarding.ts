import { useAuthStore } from '../store/auth'
import { api } from './client'
import type { CVUploadStatus } from '../types'

// uploadCV uses FormData — cannot go through the JSON api client
export async function uploadCV(file: File): Promise<CVUploadStatus> {
  const token = useAuthStore.getState().token
  const form = new FormData()
  form.append('cv', file)

  const res = await fetch('/api/onboarding/cv', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token ?? ''}` },
    body: form,
  })

  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? 'Upload failed')
  return body as CVUploadStatus
}

export function getCVStatus(): Promise<CVUploadStatus> {
  return api.get<CVUploadStatus>('/onboarding/cv/status')
}
