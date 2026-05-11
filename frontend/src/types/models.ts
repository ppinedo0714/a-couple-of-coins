export type AccountType = 'checking' | 'savings' | 'credit' | 'investment'
export type TransactionSource = 'manual' | 'csv' | 'bank'
export type ImportJobStatus = 'pending' | 'processing' | 'done' | 'failed'

export type User = {
  id: string
  email: string
  created_at: string
}

export type Account = {
  id: string
  name: string
  type: AccountType
  balance: number
  currency: string
  created_at: string
}

export type Category = {
  id: string
  name: string
  color: string | null
  parent_id: string | null
  created_at: string
}

export type Transaction = {
  id: string
  account_id: string
  category_id: string | null
  amount: number
  description: string
  merchant_name: string | null
  date: string
  source: TransactionSource
  classified: boolean
  created_at: string
}

export type ImportJob = {
  id: string
  status: ImportJobStatus
  source_type: 'csv'
  file_name: string
  rows_total: number | null
  rows_imported: number
  created_at: string
  completed_at: string | null
}

export type AccountBalanceSnapshot = {
  date: string
  account_id: string
  balance: number
}
