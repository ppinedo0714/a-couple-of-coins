import { apiFetch } from './client'
import type { Transaction } from '@/types/models'
import type {
  ClassifyResponse,
  CreateTransactionRequest,
  ListTransactionsQuery,
  ListTransactionsResponse,
  UpdateTransactionRequest,
} from '@/types/api'

function buildQuery(query: ListTransactionsQuery): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === '') continue
    params.set(key, String(value))
  }
  const s = params.toString()
  return s ? `?${s}` : ''
}

export function listTransactions(query: ListTransactionsQuery = {}) {
  return apiFetch<ListTransactionsResponse>(`/transactions${buildQuery(query)}`)
}

export function getTransaction(id: string) {
  return apiFetch<Transaction>(`/transactions/${id}`)
}

export function createTransaction(body: CreateTransactionRequest) {
  return apiFetch<Transaction>('/transactions', { method: 'POST', body })
}

export function updateTransaction(id: string, body: UpdateTransactionRequest) {
  return apiFetch<Transaction>(`/transactions/${id}`, { method: 'PUT', body })
}

export function deleteTransaction(id: string) {
  return apiFetch<void>(`/transactions/${id}`, { method: 'DELETE' })
}

export function classifyTransactions() {
  return apiFetch<ClassifyResponse>('/transactions/classify', { method: 'POST' })
}
