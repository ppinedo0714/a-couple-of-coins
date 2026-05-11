import type {
  Account,
  AccountBalanceSnapshot,
  AccountType,
  Category,
  ImportJob,
  Transaction,
  User,
} from './models'

// Auth
export type AuthCredentials = {
  email: string
  password: string
}

export type AuthResponse = {
  user: User
}

// Users
export type UpdateUserRequest = {
  email?: string
}

// Accounts
export type CreateAccountRequest = {
  name: string
  type: AccountType
  balance: number
  currency: string
}

export type UpdateAccountRequest = Partial<{
  name: string
  type: AccountType
}>

// Categories
export type CreateCategoryRequest = {
  name: string
  color?: string
  parent_id?: string
}

export type UpdateCategoryRequest = Partial<{
  name: string
  color: string
}>

// Transactions
export type ListTransactionsQuery = Partial<{
  account_id: string
  category_id: string
  from: string
  to: string
  search: string
  unclassified: boolean
  limit: number
  offset: number
}>

export type ListTransactionsResponse = {
  transactions: Transaction[]
  total: number
  limit: number
  offset: number
}

export type CreateTransactionRequest = {
  account_id: string
  category_id?: string | null
  amount: number
  description: string
  date: string
}

export type UpdateTransactionRequest = Partial<{
  account_id: string
  category_id: string | null
  amount: number
  description: string
  date: string
}>

export type ClassifyResponse = {
  classified: number
  failed: number
}

// Imports
export type CreateImportResponse = {
  job_id: string
  status: ImportJob['status']
}

// Account history
export type AccountHistoryQuery = {
  account_ids?: string[]
  from: string
  to: string
  interval?: 'day' | 'week' | 'month'
}

export type AccountHistoryResponse = {
  snapshots: AccountBalanceSnapshot[]
}

// Aliases re-exported for convenience
export type { Account, AccountBalanceSnapshot, Category, Transaction, ImportJob, User }
