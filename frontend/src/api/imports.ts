import { apiFetch } from './client'
import type { ImportJob } from '@/types/models'
import type { CreateImportResponse } from '@/types/api'

export function uploadCsv(file: File, accountId: string) {
  const form = new FormData()
  form.append('file', file)
  form.append('account_id', accountId)
  return apiFetch<CreateImportResponse>('/import/csv', { method: 'POST', body: form })
}

export function listImports() {
  return apiFetch<ImportJob[]>('/import/jobs')
}

export function getImportJob(id: string) {
  return apiFetch<ImportJob>(`/import/jobs/${id}`)
}
