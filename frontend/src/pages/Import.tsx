import { useEffect, useRef, useState, type DragEvent } from 'react'
import { toast } from 'sonner'
import { FileText, Link as LinkIcon, Upload } from 'lucide-react'
import type { ImportJob } from '@/types/models'
import { useAccounts } from '@/hooks/useAccounts'
import { useImportJob, useImports, useUploadCsv } from '@/hooks/useImports'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { EmptyState } from '@/components/shared/EmptyState'
import { PageWrapper } from '@/components/layout/PageWrapper'
import { formatDate } from '@/lib/format'
import { cn } from '@/lib/utils'

function statusVariant(status: ImportJob['status']) {
  switch (status) {
    case 'done':
      return 'income' as const
    case 'failed':
      return 'destructive' as const
    case 'pending':
    case 'processing':
      return 'default' as const
  }
}

export default function ImportPage() {
  const accountsQuery = useAccounts()
  const importsQuery = useImports()
  const uploadMut = useUploadCsv()
  const fileRef = useRef<HTMLInputElement>(null)
  const [selectedAccountId, setSelectedAccountId] = useState<string | null>(null)
  const [activeJobId, setActiveJobId] = useState<string | null>(null)
  const [dragOver, setDragOver] = useState(false)
  const announcedJobs = useRef<Set<string>>(new Set())
  const activeJobQuery = useImportJob(activeJobId)

  const accounts = accountsQuery.data ?? []
  const imports = importsQuery.data ?? []
  const accountId = selectedAccountId ?? accounts[0]?.id ?? null

  const handleFile = async (file: File) => {
    if (!accountId) {
      toast.error('Pick an account first')
      return
    }
    try {
      const res = await uploadMut.mutateAsync({ file, accountId })
      setActiveJobId(res.job_id)
      toast.success(`Uploading ${file.name}…`)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Upload failed')
    }
  }

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) void handleFile(file)
  }

  const activeJob = activeJobQuery.data

  useEffect(() => {
    if (!activeJob) return
    if (activeJob.status !== 'done' && activeJob.status !== 'failed') return
    if (announcedJobs.current.has(activeJob.id)) return
    announcedJobs.current.add(activeJob.id)
    if (activeJob.status === 'done') toast.success(`Imported ${activeJob.rows_imported} rows`)
    else toast.error('Import failed')
  }, [activeJob])

  return (
    <PageWrapper className="space-y-8">
      <div>
        <h1 className="font-serif text-3xl">Import transactions</h1>
        <p className="text-sm text-muted-foreground">
          Upload a CSV from your bank or — soon — link an account directly.
        </p>
      </div>

      <Card className="opacity-70">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2 text-base">
                <LinkIcon className="h-4 w-4" />
                Link a bank account
              </CardTitle>
              <CardDescription>
                Connect via Plaid for automatic, daily transaction sync.
              </CardDescription>
            </div>
            <Tooltip>
              <TooltipTrigger asChild>
                <span>
                  <Button disabled variant="outline" size="sm">
                    Coming soon
                  </Button>
                </span>
              </TooltipTrigger>
              <TooltipContent>Bank linking is coming in a future release.</TooltipContent>
            </Tooltip>
          </div>
        </CardHeader>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <FileText className="h-4 w-4" />
            Upload CSV
          </CardTitle>
          <CardDescription>Most banks export columns: date, description, amount.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap items-center gap-3">
            <span className="text-sm text-muted-foreground">Import into</span>
            <Select value={accountId ?? undefined} onValueChange={setSelectedAccountId}>
              <SelectTrigger className="w-56">
                <SelectValue placeholder="Pick an account" />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={a.id}>
                    {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div
            onDragOver={(e) => {
              e.preventDefault()
              setDragOver(true)
            }}
            onDragLeave={() => setDragOver(false)}
            onDrop={onDrop}
            className={cn(
              'flex flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed border-border bg-card/50 px-6 py-10 text-center transition-colors',
              dragOver && 'border-primary bg-primary/5',
            )}
          >
            <Upload className="h-6 w-6 text-muted-foreground" />
            <div className="text-sm text-foreground">Drag and drop a CSV here</div>
            <div className="text-xs text-muted-foreground">or</div>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => fileRef.current?.click()}
              disabled={uploadMut.isPending || !accountId}
            >
              {uploadMut.isPending ? <LoadingSpinner /> : null}
              Choose file
            </Button>
            <input
              ref={fileRef}
              type="file"
              accept=".csv,text/csv"
              className="hidden"
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (file) void handleFile(file)
                e.target.value = ''
              }}
            />
          </div>

          {activeJob ? (
            <div className="rounded-md border border-border bg-card p-4">
              <div className="flex items-center justify-between">
                <div className="text-sm">
                  <div className="font-medium text-foreground">{activeJob.file_name}</div>
                  <div className="text-xs text-muted-foreground">
                    {activeJob.rows_imported} of {activeJob.rows_total ?? '—'} rows
                  </div>
                </div>
                <Badge variant={statusVariant(activeJob.status)}>{activeJob.status}</Badge>
              </div>
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Import history</CardTitle>
        </CardHeader>
        <CardContent className="px-0">
          {imports.length === 0 ? (
            <div className="px-6 pb-6">
              <EmptyState
                title="No imports yet"
                description="Your CSV uploads will show up here."
              />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>File</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Rows</TableHead>
                  <TableHead>Created</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {imports.map((job) => (
                  <TableRow key={job.id}>
                    <TableCell className="text-sm font-medium">{job.file_name}</TableCell>
                    <TableCell>
                      <Badge variant={statusVariant(job.status)}>{job.status}</Badge>
                    </TableCell>
                    <TableCell className="text-right font-mono text-sm tabular-nums">
                      {job.rows_imported}
                      {job.rows_total ? ` / ${job.rows_total}` : ''}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {formatDate(job.created_at)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </PageWrapper>
  )
}
