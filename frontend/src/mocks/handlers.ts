import { http, HttpResponse, delay } from 'msw'
import { db, dbHelpers } from './db'
import type {
  Account,
  Category,
  ImportJob,
  Transaction,
} from '@/types/models'

const API = '/api/v1'

function unauthorized() {
  return HttpResponse.json({ error: 'unauthorized' }, { status: 401 })
}

function requireAuth() {
  const user = dbHelpers.authedUser()
  if (!user) return null
  return user
}

export const handlers = [
  // ----- Auth -----
  http.post(`${API}/auth/register`, async ({ request }) => {
    const body = (await request.json()) as { email?: string; password?: string }
    if (!body.email || !body.password || body.password.length < 8) {
      return HttpResponse.json({ error: 'Invalid input' }, { status: 400 })
    }
    if (db.users.some((u) => u.email.toLowerCase() === body.email!.toLowerCase())) {
      return HttpResponse.json({ error: 'Email already registered' }, { status: 409 })
    }
    const user = {
      id: dbHelpers.uuid(),
      email: body.email,
      password: body.password,
      created_at: dbHelpers.isoTimestamp(),
    }
    db.users.push(user)
    db.sessionUserId = user.id
    return HttpResponse.json(
      { user: { id: user.id, email: user.email, created_at: user.created_at } },
      { status: 201 },
    )
  }),

  http.post(`${API}/auth/login`, async ({ request }) => {
    const body = (await request.json()) as { email?: string; password?: string }
    const user = db.users.find(
      (u) => u.email.toLowerCase() === body.email?.toLowerCase() && u.password === body.password,
    )
    if (!user) return HttpResponse.json({ error: 'Invalid credentials' }, { status: 401 })
    db.sessionUserId = user.id
    return HttpResponse.json({
      user: { id: user.id, email: user.email, created_at: user.created_at },
    })
  }),

  http.post(`${API}/auth/logout`, async () => {
    db.sessionUserId = null
    return new HttpResponse(null, { status: 204 })
  }),

  http.get(`${API}/auth/oauth/:provider`, async ({ params }) => {
    // Simulate provider auto-grant: authenticate as demo user and bounce to login?oauth=success
    const provider = params.provider as string
    if (!['google', 'github'].includes(provider)) {
      return HttpResponse.json({ error: 'unknown provider' }, { status: 404 })
    }
    const demo = db.users[0]
    if (demo) db.sessionUserId = demo.id
    return HttpResponse.redirect(`${window.location.origin}/login?oauth=success`, 302)
  }),

  // ----- Users -----
  http.get(`${API}/users/me`, () => {
    const user = requireAuth()
    if (!user) return unauthorized()
    return HttpResponse.json(user)
  }),

  http.put(`${API}/users/me`, async ({ request }) => {
    const user = requireAuth()
    if (!user) return unauthorized()
    const body = (await request.json()) as { email?: string }
    const stored = db.users.find((u) => u.id === user.id)
    if (!stored) return unauthorized()
    if (body.email) {
      if (db.users.some((u) => u.id !== stored.id && u.email.toLowerCase() === body.email!.toLowerCase())) {
        return HttpResponse.json({ error: 'Email taken' }, { status: 409 })
      }
      stored.email = body.email
    }
    return HttpResponse.json({ id: stored.id, email: stored.email, created_at: stored.created_at })
  }),

  // ----- Accounts -----
  http.get(`${API}/accounts`, () => {
    if (!requireAuth()) return unauthorized()
    return HttpResponse.json(db.accounts)
  }),

  http.post(`${API}/accounts`, async ({ request }) => {
    if (!requireAuth()) return unauthorized()
    const body = (await request.json()) as Partial<Account>
    if (!body.name || !body.type) {
      return HttpResponse.json({ error: 'name and type required' }, { status: 400 })
    }
    const account: Account = {
      id: dbHelpers.uuid(),
      name: body.name,
      type: body.type,
      balance: body.balance ?? 0,
      currency: body.currency ?? 'USD',
      created_at: dbHelpers.isoTimestamp(),
    }
    db.accounts.push(account)
    return HttpResponse.json(account, { status: 201 })
  }),

  // /accounts/history must be registered before /accounts/:id so the literal path wins
  http.get(`${API}/accounts/history`, ({ request }) => {
    if (!requireAuth()) return unauthorized()
    const url = new URL(request.url)
    const accountIdsParam = url.searchParams.get('account_ids')
    const accountIds = accountIdsParam ? accountIdsParam.split(',').filter(Boolean) : null
    const from = url.searchParams.get('from') ?? ''
    const to = url.searchParams.get('to') ?? ''
    const interval = (url.searchParams.get('interval') ?? 'day') as 'day' | 'week' | 'month'

    let filtered = [...db.accountSnapshots]
    if (accountIds?.length) filtered = filtered.filter((s) => accountIds.includes(s.account_id))
    if (from) filtered = filtered.filter((s) => s.date >= from)
    if (to) filtered = filtered.filter((s) => s.date <= to)
    filtered.sort((a, b) => a.date.localeCompare(b.date))

    if (interval === 'month') {
      const byMonth = new Map<string, typeof filtered[number]>()
      for (const snap of filtered) {
        byMonth.set(`${snap.account_id}-${snap.date.slice(0, 7)}`, snap)
      }
      filtered = Array.from(byMonth.values()).sort((a, b) => a.date.localeCompare(b.date))
    }

    return HttpResponse.json({
      snapshots: filtered.map((s) => ({
        date: s.date,
        account_id: s.account_id,
        balance: s.balance,
      })),
    })
  }),

  http.get(`${API}/accounts/:id`, ({ params }) => {
    if (!requireAuth()) return unauthorized()
    const account = db.accounts.find((a) => a.id === params.id)
    if (!account) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    return HttpResponse.json(account)
  }),

  http.put(`${API}/accounts/:id`, async ({ params, request }) => {
    if (!requireAuth()) return unauthorized()
    const account = db.accounts.find((a) => a.id === params.id)
    if (!account) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    const body = (await request.json()) as Partial<Account>
    if (body.name !== undefined) account.name = body.name
    if (body.type !== undefined) account.type = body.type
    return HttpResponse.json(account)
  }),

  http.delete(`${API}/accounts/:id`, ({ params }) => {
    if (!requireAuth()) return unauthorized()
    const idx = db.accounts.findIndex((a) => a.id === params.id)
    if (idx < 0) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    if (db.transactions.some((t) => t.account_id === params.id)) {
      return HttpResponse.json({ error: 'account has transactions' }, { status: 409 })
    }
    db.accounts.splice(idx, 1)
    return new HttpResponse(null, { status: 204 })
  }),

  // ----- Categories -----
  http.get(`${API}/categories`, () => {
    if (!requireAuth()) return unauthorized()
    return HttpResponse.json(db.categories)
  }),

  http.post(`${API}/categories`, async ({ request }) => {
    if (!requireAuth()) return unauthorized()
    const body = (await request.json()) as Partial<Category> & { parent_id?: string }
    if (!body.name) return HttpResponse.json({ error: 'name required' }, { status: 400 })
    const parentId = body.parent_id ?? null
    if (parentId !== null) {
      const parent = db.categories.find((c) => c.id === parentId)
      if (!parent || parent.parent_id !== null) {
        return HttpResponse.json({ error: 'parent must be a group' }, { status: 400 })
      }
    }
    if (db.categories.some((c) => c.parent_id === parentId && c.name.toLowerCase() === body.name!.toLowerCase())) {
      return HttpResponse.json({ error: 'name already exists' }, { status: 409 })
    }
    const cat: Category = {
      id: dbHelpers.uuid(),
      name: body.name,
      color: parentId !== null ? null : (body.color ?? null),
      parent_id: parentId,
      created_at: dbHelpers.isoTimestamp(),
    }
    db.categories.push(cat)
    return HttpResponse.json(cat, { status: 201 })
  }),

  http.put(`${API}/categories/:id`, async ({ params, request }) => {
    if (!requireAuth()) return unauthorized()
    const cat = db.categories.find((c) => c.id === params.id)
    if (!cat) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    const body = (await request.json()) as Partial<Category>
    if (body.name !== undefined) cat.name = body.name
    if (body.color !== undefined && cat.parent_id === null) cat.color = body.color
    return HttpResponse.json(cat)
  }),

  http.delete(`${API}/categories/:id`, ({ params }) => {
    if (!requireAuth()) return unauthorized()
    const idx = db.categories.findIndex((c) => c.id === params.id)
    if (idx < 0) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    db.categories.forEach((c) => {
      if (c.parent_id === params.id) c.parent_id = null
    })
    db.transactions.forEach((t) => {
      if (t.category_id === params.id) t.category_id = null
    })
    db.categories.splice(idx, 1)
    return new HttpResponse(null, { status: 204 })
  }),

  // ----- Transactions -----
  http.get(`${API}/transactions`, ({ request }) => {
    if (!requireAuth()) return unauthorized()
    const url = new URL(request.url)
    const accountId = url.searchParams.get('account_id')
    const categoryId = url.searchParams.get('category_id')
    const from = url.searchParams.get('from')
    const to = url.searchParams.get('to')
    const search = url.searchParams.get('search')?.toLowerCase()
    const unclassified = url.searchParams.get('unclassified') === 'true'
    const limit = Math.min(Number(url.searchParams.get('limit') ?? 50), 200)
    const offset = Number(url.searchParams.get('offset') ?? 0)

    let filtered = [...db.transactions]
    if (accountId) filtered = filtered.filter((t) => t.account_id === accountId)
    if (categoryId) {
      const isGroup = db.categories.some((c) => c.id === categoryId && c.parent_id === null)
      if (isGroup) {
        const childIds = db.categories.filter((c) => c.parent_id === categoryId).map((c) => c.id)
        filtered = filtered.filter((t) => t.category_id === categoryId || childIds.includes(t.category_id ?? ''))
      } else {
        filtered = filtered.filter((t) => t.category_id === categoryId)
      }
    }
    if (from) filtered = filtered.filter((t) => t.date >= from)
    if (to) filtered = filtered.filter((t) => t.date <= to)
    if (search) {
      filtered = filtered.filter(
        (t) =>
          t.description.toLowerCase().includes(search) ||
          t.merchant_name?.toLowerCase().includes(search),
      )
    }
    if (unclassified) filtered = filtered.filter((t) => t.category_id === null)
    filtered.sort((a, b) => (a.date < b.date ? 1 : -1))
    const total = filtered.length
    const page = filtered.slice(offset, offset + limit)
    return HttpResponse.json({ transactions: page, total, limit, offset })
  }),

  http.post(`${API}/transactions`, async ({ request }) => {
    if (!requireAuth()) return unauthorized()
    const body = (await request.json()) as Partial<Transaction>
    if (!body.account_id || body.amount === undefined || !body.description || !body.date) {
      return HttpResponse.json({ error: 'invalid input' }, { status: 400 })
    }
    if (!db.accounts.find((a) => a.id === body.account_id)) {
      return HttpResponse.json({ error: 'account not found' }, { status: 404 })
    }
    const tx: Transaction = {
      id: dbHelpers.uuid(),
      account_id: body.account_id,
      category_id: body.category_id ?? null,
      amount: body.amount,
      description: body.description,
      merchant_name: body.merchant_name ?? null,
      date: body.date,
      source: 'manual',
      classified: body.category_id != null,
      created_at: dbHelpers.isoTimestamp(),
    }
    db.transactions.push(tx)
    const account = db.accounts.find((a) => a.id === tx.account_id)
    if (account) account.balance += tx.amount
    return HttpResponse.json(tx, { status: 201 })
  }),

  http.get(`${API}/transactions/:id`, ({ params }) => {
    if (!requireAuth()) return unauthorized()
    const tx = db.transactions.find((t) => t.id === params.id)
    if (!tx) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    return HttpResponse.json(tx)
  }),

  http.put(`${API}/transactions/:id`, async ({ params, request }) => {
    if (!requireAuth()) return unauthorized()
    const tx = db.transactions.find((t) => t.id === params.id)
    if (!tx) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    const body = (await request.json()) as Partial<Transaction>
    if (body.account_id !== undefined) tx.account_id = body.account_id
    if (body.category_id !== undefined) {
      tx.category_id = body.category_id
      tx.classified = body.category_id != null
    }
    if (body.amount !== undefined) tx.amount = body.amount
    if (body.description !== undefined) tx.description = body.description
    if (body.date !== undefined) tx.date = body.date
    return HttpResponse.json(tx)
  }),

  http.delete(`${API}/transactions/:id`, ({ params }) => {
    if (!requireAuth()) return unauthorized()
    const idx = db.transactions.findIndex((t) => t.id === params.id)
    if (idx < 0) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    const tx = db.transactions[idx]!
    const account = db.accounts.find((a) => a.id === tx.account_id)
    if (account) account.balance -= tx.amount
    db.transactions.splice(idx, 1)
    return new HttpResponse(null, { status: 204 })
  }),

  http.post(`${API}/transactions/classify`, () => {
    if (!requireAuth()) return unauthorized()
    let classified = 0
    for (const tx of db.transactions) {
      if (tx.classified || tx.category_id) continue
      tx.category_id = 'cat-groceries'
      tx.classified = true
      tx.merchant_name = tx.merchant_name ?? 'Auto-detected'
      classified++
    }
    return HttpResponse.json({ classified, failed: 0 })
  }),

  // ----- Imports -----
  http.post(`${API}/import/csv`, async ({ request }) => {
    if (!requireAuth()) return unauthorized()
    const form = await request.formData()
    const file = form.get('file') as File | null
    const accountId = form.get('account_id') as string | null
    if (!file) return HttpResponse.json({ error: 'file required' }, { status: 400 })
    if (!accountId || !db.accounts.find((a) => a.id === accountId)) {
      return HttpResponse.json({ error: 'invalid account' }, { status: 400 })
    }
    const job: ImportJob = {
      id: dbHelpers.uuid(),
      status: 'pending',
      source_type: 'csv',
      file_name: file.name,
      rows_total: null,
      rows_imported: 0,
      created_at: dbHelpers.isoTimestamp(),
      completed_at: null,
    }
    db.imports.unshift(job)

    // Simulate background progression
    const rows = 12 + Math.floor(Math.random() * 30)
    setTimeout(() => {
      job.status = 'processing'
      job.rows_total = rows
    }, 1500)
    setTimeout(() => {
      job.status = 'done'
      job.rows_imported = rows
      job.completed_at = dbHelpers.isoTimestamp()
      // Synthesize a few imported transactions
      const merchants = ['STORE A', 'STORE B', 'STORE C', 'CAFE D']
      for (let i = 0; i < Math.min(rows, 5); i++) {
        const date = new Date()
        date.setDate(date.getDate() - i)
        db.transactions.push({
          id: dbHelpers.uuid(),
          account_id: accountId,
          category_id: null,
          amount: -(5 + Math.random() * 50),
          description: `${merchants[i % merchants.length]!} ${file.name}`,
          merchant_name: null,
          date: dbHelpers.isoDate(date),
          source: 'csv',
          classified: false,
          created_at: dbHelpers.isoTimestamp(),
        })
      }
    }, 4000)

    return HttpResponse.json({ job_id: job.id, status: job.status }, { status: 202 })
  }),

  http.get(`${API}/import/jobs`, () => {
    if (!requireAuth()) return unauthorized()
    return HttpResponse.json(db.imports)
  }),

  http.get(`${API}/import/jobs/:id`, ({ params }) => {
    if (!requireAuth()) return unauthorized()
    const job = db.imports.find((j) => j.id === params.id)
    if (!job) return HttpResponse.json({ error: 'not found' }, { status: 404 })
    return HttpResponse.json(job)
  }),

  // ----- Health -----
  http.get(`${API}/health`, async () => {
    await delay(50)
    return HttpResponse.json({ status: 'ok', version: '1.0.0' })
  }),
]
