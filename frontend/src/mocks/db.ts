import type { Account, Category, ImportJob, Transaction, User } from '@/types/models'

type StoredUser = User & { password: string }

type AccountBalanceSnapshot = {
  id: string
  account_id: string
  balance: number
  date: string
  created_at: string
}

type DbState = {
  users: StoredUser[]
  accounts: Account[]
  categories: Category[]
  transactions: Transaction[]
  imports: ImportJob[]
  accountSnapshots: AccountBalanceSnapshot[]
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
    // Groups (parent_id: null)
    { id: 'grp-income', name: 'Income', color: '#26A69A', parent_id: null, created_at: isoTimestamp() },
    { id: 'grp-finance', name: 'Finance', color: '#5C6BC0', parent_id: null, created_at: isoTimestamp() },
    { id: 'grp-housing', name: 'Housing', color: '#7E57C2', parent_id: null, created_at: isoTimestamp() },
    { id: 'grp-food', name: 'Food & Drink', color: '#66BB6A', parent_id: null, created_at: isoTimestamp() },
    { id: 'grp-health', name: 'Health & Wellness', color: '#EC407A', parent_id: null, created_at: isoTimestamp() },
    { id: 'grp-entertainment', name: 'Entertainment', color: '#AB47BC', parent_id: null, created_at: isoTimestamp() },
    { id: 'grp-travel', name: 'Travel', color: '#42A5F5', parent_id: null, created_at: isoTimestamp() },
    { id: 'grp-other', name: 'Other', color: '#78909C', parent_id: null, created_at: isoTimestamp() },
    // Income sub-categories
    { id: 'cat-salary', name: 'Salary', color: null, parent_id: 'grp-income', created_at: isoTimestamp() },
    { id: 'cat-freelance', name: 'Freelance', color: null, parent_id: 'grp-income', created_at: isoTimestamp() },
    { id: 'cat-bonus', name: 'Bonus', color: null, parent_id: 'grp-income', created_at: isoTimestamp() },
    // Finance sub-categories
    { id: 'cat-savings', name: 'Savings', color: null, parent_id: 'grp-finance', created_at: isoTimestamp() },
    { id: 'cat-investments', name: 'Investments', color: null, parent_id: 'grp-finance', created_at: isoTimestamp() },
    { id: 'cat-donations', name: 'Donations', color: null, parent_id: 'grp-finance', created_at: isoTimestamp() },
    // Housing sub-categories
    { id: 'cat-rent', name: 'Rent', color: null, parent_id: 'grp-housing', created_at: isoTimestamp() },
    { id: 'cat-electricity', name: 'Electricity', color: null, parent_id: 'grp-housing', created_at: isoTimestamp() },
    { id: 'cat-internet', name: 'Internet', color: null, parent_id: 'grp-housing', created_at: isoTimestamp() },
    // Food & Drink sub-categories
    { id: 'cat-groceries', name: 'Groceries', color: null, parent_id: 'grp-food', created_at: isoTimestamp() },
    { id: 'cat-dining', name: 'Dining Out', color: null, parent_id: 'grp-food', created_at: isoTimestamp() },
    // Health & Wellness sub-categories
    { id: 'cat-medical', name: 'Medical', color: null, parent_id: 'grp-health', created_at: isoTimestamp() },
    { id: 'cat-dental', name: 'Dental', color: null, parent_id: 'grp-health', created_at: isoTimestamp() },
    { id: 'cat-gym', name: 'Gym', color: null, parent_id: 'grp-health', created_at: isoTimestamp() },
    // Entertainment sub-categories
    { id: 'cat-sports', name: 'Sports', color: null, parent_id: 'grp-entertainment', created_at: isoTimestamp() },
    { id: 'cat-books', name: 'Books', color: null, parent_id: 'grp-entertainment', created_at: isoTimestamp() },
    { id: 'cat-streaming', name: 'Streaming', color: null, parent_id: 'grp-entertainment', created_at: isoTimestamp() },
    // Travel sub-categories
    { id: 'cat-gas', name: 'Gas', color: null, parent_id: 'grp-travel', created_at: isoTimestamp() },
    { id: 'cat-flights', name: 'Flights', color: null, parent_id: 'grp-travel', created_at: isoTimestamp() },
    { id: 'cat-hotels', name: 'Hotels', color: null, parent_id: 'grp-travel', created_at: isoTimestamp() },
  ]

  const merchants: Array<{ desc: string; merchant: string; cat: string; amount: number }> = [
    // Income
    { desc: 'FREELANCE PAYMENT', merchant: 'Client Co', cat: 'cat-freelance', amount: 1200 },
    // Finance
    { desc: 'ALLY SAVINGS TRANSFER', merchant: 'Ally Bank', cat: 'cat-savings', amount: -500 },
    { desc: 'VANGUARD AUTO-INVEST', merchant: 'Vanguard', cat: 'cat-investments', amount: -300 },
    { desc: 'RED CROSS DONATION', merchant: 'Red Cross', cat: 'cat-donations', amount: -50 },
    // Housing
    { desc: 'CON ED PAYMENT', merchant: 'Con Edison', cat: 'cat-electricity', amount: -82.5 },
    { desc: 'COMCAST XFINITY', merchant: 'Xfinity', cat: 'cat-internet', amount: -79.99 },
    // Food & Drink
    { desc: 'WHOLE FOODS MARKET', merchant: 'Whole Foods', cat: 'cat-groceries', amount: -68.42 },
    { desc: 'TRADER JOES #482', merchant: 'Trader Joes', cat: 'cat-groceries', amount: -42.18 },
    { desc: 'COSTCO WHSE 1083', merchant: 'Costco', cat: 'cat-groceries', amount: -132.7 },
    { desc: 'CHIPOTLE 0421', merchant: 'Chipotle', cat: 'cat-dining', amount: -13.85 },
    { desc: 'STARBUCKS STORE 9921', merchant: 'Starbucks', cat: 'cat-dining', amount: -6.45 },
    { desc: 'BLUE BOTTLE COFFEE', merchant: 'Blue Bottle', cat: 'cat-dining', amount: -7.5 },
    // Health & Wellness
    { desc: 'PLANET FITNESS', merchant: 'Planet Fitness', cat: 'cat-gym', amount: -24.99 },
    { desc: 'DR SMITH MEDICAL', merchant: 'Dr. Smith', cat: 'cat-medical', amount: -150 },
    { desc: 'BRIGHT DENTAL', merchant: 'Bright Dental', cat: 'cat-dental', amount: -200 },
    // Entertainment
    { desc: 'NETFLIX.COM', merchant: 'Netflix', cat: 'cat-streaming', amount: -15.49 },
    { desc: 'SPOTIFY USA', merchant: 'Spotify', cat: 'cat-streaming', amount: -10.99 },
    { desc: 'AMAZON BOOKS', merchant: 'Amazon', cat: 'cat-books', amount: -16.99 },
    { desc: 'NBA TICKETS', merchant: 'NBA', cat: 'cat-sports', amount: -89 },
    // Travel
    { desc: 'SHELL OIL 5713', merchant: 'Shell', cat: 'cat-gas', amount: -54.1 },
    { desc: 'DELTA AIRLINES', merchant: 'Delta', cat: 'cat-flights', amount: -342 },
    { desc: 'MARRIOTT HOTEL', merchant: 'Marriott', cat: 'cat-hotels', amount: -189 },
    { desc: 'UBER TRIP', merchant: 'Uber', cat: 'cat-gas', amount: -18.5 },
    // Other
    { desc: 'MISC EXPENSE', merchant: 'Various', cat: 'grp-other', amount: -45 },
  ]

  const transactions: Transaction[] = []

  // Random transactions over the last 365 days (covers this year + last year)
  for (let dayOffset = 0; dayOffset < 365; dayOffset++) {
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

  // Monthly recurring: salary + rent for each of the last 13 months
  for (let monthOffset = 0; monthOffset <= 12; monthOffset++) {
    const firstOfMonth = new Date(now.getFullYear(), now.getMonth() - monthOffset, 1)
    const dateStr = isoDate(firstOfMonth)
    const ts = firstOfMonth.toISOString()
    transactions.push({
      id: uuid(),
      account_id: 'acc-checking',
      category_id: 'cat-salary',
      amount: 4250,
      description: 'PAYROLL DEPOSIT',
      merchant_name: 'Payroll',
      date: dateStr,
      source: 'csv',
      classified: true,
      created_at: ts,
    })
    transactions.push({
      id: uuid(),
      account_id: 'acc-checking',
      category_id: 'cat-rent',
      amount: -1850,
      description: 'RENT PAYMENT',
      merchant_name: 'Landlord',
      date: dateStr,
      source: 'manual',
      classified: true,
      created_at: ts,
    })
  }

  // Two unclassified on today
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

  // Generate ~70 weekly balance snapshots per account from Jan 2025 to present
  const accountSnapshots: AccountBalanceSnapshot[] = []
  const snapshotStart = new Date('2025-01-06')
  const snapshotEnd = new Date()
  let weekIndex = 0
  for (let d = new Date(snapshotStart); d <= snapshotEnd; d.setDate(d.getDate() + 7)) {
    const snap = new Date(d)
    const dateStr = isoDate(snap)
    const ts = isoTimestamp(snap)

    // Chase Checking: starts ~$1800, slow growth with a monthly paycheck spike
    const checkingBase = 1800 + weekIndex * 9.5
    const checkingCycle = 500 * Math.sin(((weekIndex % 4) / 4) * 2 * Math.PI)
    const checking = Math.round((checkingBase + checkingCycle) * 100) / 100

    // Ally Savings: steady growth from $8000 toward $12500
    const savings = Math.round((8000 + weekIndex * 64.3) * 100) / 100

    // Amex Gold: sawtooth from $0 to −$1200, paid off each month
    const credit = Math.round(-(((weekIndex % 4) / 3) * 1200) * 100) / 100

    accountSnapshots.push(
      { id: uuid(), account_id: 'acc-checking', balance: checking, date: dateStr, created_at: ts },
      { id: uuid(), account_id: 'acc-savings', balance: savings, date: dateStr, created_at: ts },
      { id: uuid(), account_id: 'acc-credit', balance: credit, date: dateStr, created_at: ts },
    )
    weekIndex++
  }

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
    accountSnapshots,
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
    db.accountSnapshots = fresh.accountSnapshots
    db.sessionUserId = null
  },
}
