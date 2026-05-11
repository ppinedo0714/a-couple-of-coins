import { apiFetch } from './client'
import type { Account } from '@/types/models'
import type { AccountHistoryQuery, AccountHistoryResponse, CreateAccountRequest, UpdateAccountRequest } from '@/types/api'

export function listAccounts() {
  return apiFetch<Account[]>('/accounts')
}

export function getAccount(id: string) {
  return apiFetch<Account>(`/accounts/${id}`)
}

export function createAccount(body: CreateAccountRequest) {
  return apiFetch<Account>('/accounts', { method: 'POST', body })
}

export function updateAccount(id: string, body: UpdateAccountRequest) {
  return apiFetch<Account>(`/accounts/${id}`, { method: 'PUT', body })
}

export function deleteAccount(id: string) {
  return apiFetch<void>(`/accounts/${id}`, { method: 'DELETE' })
}

export function getAccountHistory(query: AccountHistoryQuery) {
  const params = new URLSearchParams()
  if (query.account_ids?.length) params.set('account_ids', query.account_ids.join(','))
  params.set('from', query.from)
  params.set('to', query.to)
  if (query.interval) params.set('interval', query.interval)
  return apiFetch<AccountHistoryResponse>(`/accounts/history?${params}`)
}
