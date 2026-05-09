import { useState } from 'react'
import { Plus } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import type { Category } from '@/types/models'
import {
  useCategories,
  useCreateCategory,
  useDeleteCategory,
  useUpdateCategory,
} from '@/hooks/useCategories'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { LoadingScreen } from '@/components/shared/LoadingSpinner'
import { EmptyState } from '@/components/shared/EmptyState'

const schema = z.object({
  name: z.string().min(1, 'Required'),
  color: z.string().regex(/^#[0-9a-fA-F]{6}$/, 'Pick a color'),
})
type Values = z.infer<typeof schema>

const DEFAULT_COLOR = '#7E57C2'

function CategoryDialog({
  open,
  onOpenChange,
  category,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  category?: Category
}) {
  const create = useCreateCategory()
  const update = useUpdateCategory()
  const remove = useDeleteCategory()
  const isEdit = !!category

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: category?.name ?? '',
      color: category?.color ?? DEFAULT_COLOR,
    },
    values: category
      ? { name: category.name, color: category.color ?? DEFAULT_COLOR }
      : undefined,
  })

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      if (isEdit && category) {
        await update.mutateAsync({ id: category.id, body: values })
        toast.success('Category updated')
      } else {
        await create.mutateAsync(values)
        toast.success('Category created')
      }
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not save')
    }
  })

  const onDelete = async () => {
    if (!category) return
    try {
      await remove.mutateAsync(category.id)
      toast.success('Category deleted')
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not delete')
    }
  }

  const pending = create.isPending || update.isPending || remove.isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit category' : 'New category'}</DialogTitle>
          {!isEdit ? (
            <DialogDescription>Give it a name and pick a color.</DialogDescription>
          ) : null}
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="cat-name">Name</Label>
            <Input id="cat-name" {...form.register('name')} />
            {form.formState.errors.name ? (
              <p className="text-xs text-destructive">{form.formState.errors.name.message}</p>
            ) : null}
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="cat-color">Color</Label>
            <div className="flex items-center gap-3">
              <input
                id="cat-color"
                type="color"
                className="h-10 w-14 cursor-pointer rounded-md border border-border bg-card"
                {...form.register('color')}
              />
              <Input className="flex-1" {...form.register('color')} />
            </div>
            {form.formState.errors.color ? (
              <p className="text-xs text-destructive">{form.formState.errors.color.message}</p>
            ) : null}
          </div>
          <DialogFooter className="sm:justify-between">
            {isEdit ? (
              <Button type="button" variant="ghost" className="text-destructive" onClick={onDelete}>
                Delete
              </Button>
            ) : (
              <span />
            )}
            <div className="flex gap-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={pending}>
                {isEdit ? 'Save' : 'Create'}
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export function CategoriesTab() {
  const categoriesQuery = useCategories()
  const [editing, setEditing] = useState<Category | null>(null)
  const [creating, setCreating] = useState(false)

  if (categoriesQuery.isLoading) return <LoadingScreen />

  const categories = categoriesQuery.data ?? []

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          Categories help group transactions on the dashboard. Deleting a category leaves its
          transactions uncategorized.
        </p>
        <Button size="sm" onClick={() => setCreating(true)}>
          <Plus className="h-4 w-4" />
          New
        </Button>
      </div>

      {categories.length === 0 ? (
        <EmptyState
          title="No categories yet"
          description="Add a few to label your spend."
          action={<Button onClick={() => setCreating(true)}>Create category</Button>}
        />
      ) : (
        <div className="flex flex-wrap gap-2">
          {categories.map((c) => (
            <button
              key={c.id}
              type="button"
              onClick={() => setEditing(c)}
              className="group flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1.5 text-sm transition-colors hover:bg-muted"
            >
              <span
                className="h-2.5 w-2.5 rounded-full"
                style={{ background: c.color ?? 'var(--muted-foreground)' }}
              />
              <span>{c.name}</span>
            </button>
          ))}
        </div>
      )}

      <CategoryDialog open={creating} onOpenChange={setCreating} />
      <CategoryDialog
        open={!!editing}
        onOpenChange={(open) => {
          if (!open) setEditing(null)
        }}
        category={editing ?? undefined}
      />
    </div>
  )
}
