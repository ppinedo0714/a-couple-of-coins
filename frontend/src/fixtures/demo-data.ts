import type { Account, Category, Transaction } from '@/types/models'

const today = new Date()
const isoDate = (offsetDays: number) => {
  const d = new Date(today)
  d.setDate(d.getDate() - offsetDays)
  return d.toISOString().slice(0, 10)
}
const ts = '2024-01-01T00:00:00Z'

export const demoAccounts: Account[] = [
  {
    id: 'demo-checking',
    name: 'Everyday Checking',
    type: 'checking',
    balance: 3214.5,
    currency: 'USD',
    created_at: ts,
  },
  {
    id: 'demo-savings',
    name: 'Vacation Fund',
    type: 'savings',
    balance: 8200,
    currency: 'USD',
    created_at: ts,
  },
  {
    id: 'demo-credit',
    name: 'Travel Card',
    type: 'credit',
    balance: -612.18,
    currency: 'USD',
    created_at: ts,
  },
]

export const demoCategories: Category[] = [
  { id: 'demo-groceries', name: 'Groceries', color: '#4CAF50', created_at: ts },
  { id: 'demo-rent', name: 'Rent', color: '#7E57C2', created_at: ts },
  { id: 'demo-transport', name: 'Transport', color: '#42A5F5', created_at: ts },
  { id: 'demo-dining', name: 'Dining', color: '#FF7043', created_at: ts },
  { id: 'demo-fun', name: 'Entertainment', color: '#EC407A', created_at: ts },
  { id: 'demo-salary', name: 'Salary', color: '#26A69A', created_at: ts },
]

const sampleTx = (
  i: number,
  amount: number,
  description: string,
  merchant: string,
  category: string,
  account: string,
): Transaction => ({
  id: `demo-tx-${i}`,
  account_id: account,
  category_id: category,
  amount,
  description,
  merchant_name: merchant,
  date: isoDate(i),
  source: 'csv',
  classified: true,
  created_at: ts,
})

export const demoTransactions: Transaction[] = [
  sampleTx(0, 4500, 'PAYROLL DEPOSIT', 'Acme Co', 'demo-salary', 'demo-checking'),
  sampleTx(1, -1850, 'RENT PAYMENT', 'Landlord', 'demo-rent', 'demo-checking'),
  sampleTx(2, -64.32, 'WHOLE FOODS', 'Whole Foods', 'demo-groceries', 'demo-credit'),
  sampleTx(3, -22.5, 'UBER TRIP', 'Uber', 'demo-transport', 'demo-credit'),
  sampleTx(4, -38.75, 'CHIPOTLE 0421', 'Chipotle', 'demo-dining', 'demo-credit'),
  sampleTx(5, -15.49, 'NETFLIX', 'Netflix', 'demo-fun', 'demo-checking'),
  sampleTx(6, -9.5, 'BLUE BOTTLE', 'Blue Bottle', 'demo-dining', 'demo-credit'),
  sampleTx(7, -52.18, 'COSTCO', 'Costco', 'demo-groceries', 'demo-credit'),
  sampleTx(8, -82.5, 'CON ED', 'Con Edison', 'demo-fun', 'demo-checking'),
  sampleTx(9, -28.5, 'AMC THEATRE', 'AMC', 'demo-fun', 'demo-credit'),
  sampleTx(10, -41.62, 'TRADER JOES', 'Trader Joes', 'demo-groceries', 'demo-credit'),
  sampleTx(11, -75, 'VERIZON', 'Verizon', 'demo-fun', 'demo-checking'),
  sampleTx(12, -18.4, 'LYFT RIDE', 'Lyft', 'demo-transport', 'demo-credit'),
  sampleTx(13, -54.1, 'SHELL', 'Shell', 'demo-transport', 'demo-checking'),
  sampleTx(14, -12.6, 'STARBUCKS', 'Starbucks', 'demo-dining', 'demo-credit'),
]
