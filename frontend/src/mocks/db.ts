import type { Account, Category, ImportJob, Transaction, User } from '@/types/models'

type StoredUser = User & { password: string }

type DbState = {
  users: StoredUser[]
  accounts: Account[]
  categories: Category[]
  transactions: Transaction[]
  imports: ImportJob[]
  sessionUserId: string | null
}

function uuid(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
  return Math.random().toString(36).slice(2) + Date.now().toString(36)
}

function isoDate(date: Date): string {
  return date.toISOString().slice(0, 10)
}

function isoTimestamp(date = new Date()): string {
  return date.toISOString()
}

function seedDb(): DbState {
  const userId = 'demo-user-id'
  const now = new Date()

  const accounts: Account[] = [
    {
      id: 'acc-checking',
      name: 'Chase Checking',
      type: 'checking',
      balance: 2450.32,
      currency: 'USD',
      created_at: isoTimestamp(),
    },
    {
      id: 'acc-savings',
      name: 'Ally Savings',
      type: 'savings',
      balance: 12500,
      currency: 'USD',
      created_at: isoTimestamp(),
    },
    {
      id: 'acc-credit',
      name: 'Amex Gold',
      type: 'credit',
      balance: -842.15,
      currency: 'USD',
      created_at: isoTimestamp(),
    },
  ]

  const categories: Category[] = [
    { id: 'cat-groceries', name: 'Groceries', color: '#4CAF50', created_at: isoTimestamp() },
    { id: 'cat-rent', name: 'Rent', color: '#7E57C2', created_at: isoTimestamp() },
    { id: 'cat-transport', name: 'Transport', color: '#42A5F5', created_at: isoTimestamp() },
    { id: 'cat-dining', name: 'Dining', color: '#FF7043', created_at: isoTimestamp() },
    { id: 'cat-entertainment', name: 'Entertainment', color: '#EC407A', created_at: isoTimestamp() },
    { id: 'cat-salary', name: 'Salary', color: '#26A69A', created_at: isoTimestamp() },
    { id: 'cat-utilities', name: 'Utilities', color: '#FFA726', created_at: isoTimestamp() },
  ]

  const merchants: Array<{ desc: string; merchant: string; cat: string; amount: number }> = [
    { desc: 'WHOLE FOODS MARKET', merchant: 'Whole Foods', cat: 'cat-groceries', amount: -68.42 },
    { desc: 'TRADER JOES #482', merchant: 'Trader Joes', cat: 'cat-groceries', amount: -42.18 },
    { desc: 'UBER TRIP', merchant: 'Uber', cat: 'cat-transport', amount: -18.5 },
    { desc: 'LYFT *RIDE', merchant: 'Lyft', cat: 'cat-transport', amount: -22.75 },
    { desc: 'SHELL OIL 5713', merchant: 'Shell', cat: 'cat-transport', amount: -54.1 },
    { desc: 'CHIPOTLE 0421', merchant: 'Chipotle', cat: 'cat-dining', amount: -13.85 },
    { desc: 'STARBUCKS STORE 9921', merchant: 'Starbucks', cat: 'cat-dining', amount: -6.45 },
    { desc: 'BLUE BOTTLE COFFEE', merchant: 'Blue Bottle', cat: 'cat-dining', amount: -7.5 },
    { desc: 'NETFLIX.COM', merchant: 'Netflix', cat: 'cat-entertainment', amount: -15.49 },
    { desc: 'SPOTIFY USA', merchant: 'Spotify', cat: 'cat-entertainment', amount: -10.99 },
    { desc: 'CON ED PAYMENT', merchant: 'Con Edison', cat: 'cat-utilities', amount: -82.5 },
    { desc: 'VERIZON WIRELESS', merchant: 'Verizon', cat: 'cat-utilities', amount: -75.0 },
    { desc: 'AMC THEATRES', merchant: 'AMC', cat: 'cat-entertainment', amount: -28.5 },
    { desc: 'COSTCO WHSE 1083', merchant: 'Costco', cat: 'cat-groceries', amount: -132.7 },
  ]

  const transactions: Transaction[] = []
  for (let dayOffset = 0; dayOffset < 45; dayOffset++) {
    const date = new Date(now)
    date.setDate(date.getDate() - dayOffset)
    const txCount = Math.floor(Math.random() * 3)
    for (let i = 0; i < txCount; i++) {
      const m = merchants[Math.floor(Math.random() * merchants.length)]!
      transactions.push({
        id: uuid(),
        account_id: Math.random() > 0.4 ? 'acc-credit' : 'acc-checking',
        category_id: m.cat,
        amount: m.amount * (0.85 + Math.random() * 0.3),
        description: m.desc,
        merchant_name: m.merchant,
        date: isoDate(date),
        source: 'csv',
        classified: true,
        created_at: date.toISOString(),
      })
    }
  }
  // Two unclassified
  transactions.push({
    id: uuid(),
    account_id: 'acc-checking',
    category_id: null,
    amount: -39.99,
    description: 'SQ *ARTISAN BAKERY',
    merchant_name: null,
    date: isoDate(now),
    source: 'csv',
    classified: false,
    created_at: now.toISOString(),
  })
  // Salary deposit
  transactions.push({
    id: uuid(),
    account_id: 'acc-checking',
    category_id: 'cat-salary',
    amount: 4250,
    description: 'PAYROLL DEPOSIT',
    merchant_name: 'Payroll',
    date: isoDate(new Date(now.getFullYear(), now.getMonth(), 1)),
    source: 'csv',
    classified: true,
    created_at: now.toISOString(),
  })
  // Rent
  transactions.push({
    id: uuid(),
    account_id: 'acc-checking',
    category_id: 'cat-rent',
    amount: -1850,
    description: 'RENT PAYMENT',
    merchant_name: 'Landlord',
    date: isoDate(new Date(now.getFullYear(), now.getMonth(), 1)),
    source: 'manual',
    classified: true,
    created_at: now.toISOString(),
  })

  return {
    users: [
      {
        id: userId,
        email: 'demo@example.com',
        password: 'password123',
        created_at: isoTimestamp(),
      },
    ],
    accounts,
    categories,
    transactions,
    imports: [],
    sessionUserId: null,
  }
}

export const db: DbState = seedDb()

export const dbHelpers = {
  uuid,
  isoTimestamp,
  isoDate,
  authedUser(): User | null {
    if (!db.sessionUserId) return null
    const u = db.users.find((x) => x.id === db.sessionUserId)
    if (!u) return null
    const { password: _password, ...rest } = u
    void _password
    return rest
  },
  resetSeed() {
    const fresh = seedDb()
    db.users = fresh.users
    db.accounts = fresh.accounts
    db.categories = fresh.categories
    db.transactions = fresh.transactions
    db.imports = fresh.imports
    db.sessionUserId = null
  },
}
