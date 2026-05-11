import { useState } from 'react'
import { ChevronDown, ChevronRight, Pencil, Plus } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import type { Category } from '@/types/models'
import {
  useCategories,
  useCreateCategory,
  useDeleteCategory,
  useGroupedCategories,
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

const groupSchema = z.object({
  name: z.string().min(1, 'Required'),
  color: z.string().regex(/^#[0-9a-fA-F]{6}$/, 'Pick a color'),
})
type GroupValues = z.infer<typeof groupSchema>

const categorySchema = z.object({
  name: z.string().min(1, 'Required'),
})
type CategoryValues = z.infer<typeof categorySchema>

const DEFAULT_COLOR = '#7E57C2'

function GroupDialog({
  open,
  onOpenChange,
  group,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  group?: Category
}) {
  const create = useCreateCategory()
  const update = useUpdateCategory()
  const remove = useDeleteCategory()
  const isEdit = !!group

  const form = useForm<GroupValues>({
    resolver: zodResolver(groupSchema),
    defaultValues: { name: group?.name ?? '', color: group?.color ?? DEFAULT_COLOR },
    values: group ? { name: group.name, color: group.color ?? DEFAULT_COLOR } : undefined,
  })

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      if (isEdit && group) {
        await update.mutateAsync({ id: group.id, body: values })
        toast.success('Group updated')
      } else {
        await create.mutateAsync(values)
        toast.success('Group created')
      }
      onOpenChange(false)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not save')
    }
  })

  const onDelete = async () => {
    if (!group) return
    try {
      await remove.mutateAsync(group.id)
      toast.success('Group deleted')
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
          <DialogTitle>{isEdit ? 'Edit group' : 'New group'}</DialogTitle>
          {!isEdit && (
            <DialogDescription>
              Groups are broad labels like Food or Entertainment.
            </DialogDescription>
          )}
        </DialogHeader>
        {isEdit && (
          <p className="text-xs text-muted-foreground">
            Deleting a group moves its categories to the top level as new groups.
          </p>
        )}
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="group-name">Name</Label>
            <Input id="group-name" {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-destructive">{form.formState.errors.name.message}</p>
            )}
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="group-color">Color</Label>
            <div className="flex items-center gap-3">
              <input
                id="group-color"
                type="color"
                className="h-10 w-14 cursor-pointer rounded-md border border-border bg-card"
                {...form.register('color')}
              />
              <Input className="flex-1" {...form.register('color')} />
            </div>
            {form.formState.errors.color && (
              <p className="text-xs text-destructive">{form.formState.errors.color.message}</p>
            )}
          </div>
          <DialogFooter className="sm:justify-between">
            {isEdit ? (
              <Button
                type="button"
                variant="ghost"
                className="text-destructive"
                onClick={onDelete}
                disabled={pending}
              >
                Delete group
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

function CategoryItemDialog({
  open,
  onOpenChange,
  category,
  parentGroup,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  category?: Category
  parentGroup?: Category
}) {
  const create = useCreateCategory()
  const update = useUpdateCategory()
  const remove = useDeleteCategory()
  const isEdit = !!category

  const form = useForm<CategoryValues>({
    resolver: zodResolver(categorySchema),
    defaultValues: { name: category?.name ?? '' },
    values: category ? { name: category.name } : undefined,
  })

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      if (isEdit && category) {
        await update.mutateAsync({ id: category.id, body: { name: values.name } })
        toast.success('Category updated')
      } else {
        await create.mutateAsync({ name: values.name, parent_id: parentGroup?.id })
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
          {!isEdit && parentGroup && (
            <DialogDescription>
              Adding a category under <strong>{parentGroup.name}</strong>.
            </DialogDescription>
          )}
        </DialogHeader>
        {isEdit && (
          <p className="text-xs text-muted-foreground">
            Deleting this category will leave its transactions uncategorized.
          </p>
        )}
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="cat-name">Name</Label>
            <Input id="cat-name" {...form.register('name')} />
            {form.formState.errors.name && (
              <p className="text-xs text-destructive">{form.formState.errors.name.message}</p>
            )}
          </div>
          {parentGroup && (
            <div className="flex items-center gap-2 rounded-md border border-border bg-muted/50 px-3 py-2">
              <span
                className="h-2.5 w-2.5 flex-shrink-0 rounded-full"
                style={{ background: parentGroup.color ?? 'var(--muted-foreground)' }}
              />
              <span className="text-xs text-muted-foreground">
                Inherits color from{' '}
                <span className="font-medium text-foreground">{parentGroup.name}</span>
              </span>
            </div>
          )}
          <DialogFooter className="sm:justify-between">
            {isEdit ? (
              <Button
                type="button"
                variant="ghost"
                className="text-destructive"
                onClick={onDelete}
                disabled={pending}
              >
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
  const { groups, categoriesByGroupId } = useGroupedCategories()
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())

  const [groupDialogOpen, setGroupDialogOpen] = useState(false)
  const [editingGroup, setEditingGroup] = useState<Category | null>(null)

  const [catDialogOpen, setCatDialogOpen] = useState(false)
  const [catDialogParentGroup, setCatDialogParentGroup] = useState<Category | null>(null)
  const [editingCategory, setEditingCategory] = useState<Category | null>(null)

  if (categoriesQuery.isLoading) return <LoadingScreen />

  const toggleGroup = (id: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const openCreateGroup = () => {
    setEditingGroup(null)
    setGroupDialogOpen(true)
  }

  const openEditGroup = (group: Category) => {
    setEditingGroup(group)
    setGroupDialogOpen(true)
  }

  const openCreateCategory = (group: Category) => {
    setEditingCategory(null)
    setCatDialogParentGroup(group)
    setCatDialogOpen(true)
  }

  const openEditCategory = (cat: Category) => {
    const parent = groups.find((g) => g.id === cat.parent_id) ?? null
    setEditingCategory(cat)
    setCatDialogParentGroup(parent)
    setCatDialogOpen(true)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          Groups are broad labels like Food or Entertainment. Add specific categories inside each
          group for more detailed tracking.
        </p>
        <Button size="sm" onClick={openCreateGroup}>
          <Plus className="h-4 w-4" />
          New group
        </Button>
      </div>

      {groups.length === 0 ? (
        <EmptyState
          title="No groups yet"
          description="Create a group to start organizing your transactions."
          action={<Button onClick={openCreateGroup}>Create group</Button>}
        />
      ) : (
        <div className="space-y-1.5">
          {groups.map((group) => {
            const children = categoriesByGroupId[group.id] ?? []
            const isExpanded = expandedGroups.has(group.id)
            const ChevronIcon = isExpanded ? ChevronDown : ChevronRight

            return (
              <div key={group.id} className="overflow-hidden rounded-lg border border-border bg-card">
                <div className="flex items-center gap-2 px-3 py-2.5">
                  <button
                    type="button"
                    onClick={() => toggleGroup(group.id)}
                    className="flex flex-1 items-center gap-2 text-left"
                  >
                    <ChevronIcon className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
                    <span
                      className="h-2.5 w-2.5 flex-shrink-0 rounded-full"
                      style={{ background: group.color ?? 'var(--muted-foreground)' }}
                    />
                    <span className="text-sm font-medium">{group.name}</span>
                    {children.length > 0 && (
                      <span className="text-xs text-muted-foreground">
                        {children.length} {children.length === 1 ? 'category' : 'categories'}
                      </span>
                    )}
                  </button>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-7 px-2 text-muted-foreground hover:text-foreground"
                    onClick={() => openEditGroup(group)}
                  >
                    <Pencil className="h-3.5 w-3.5" />
                  </Button>
                </div>

                {isExpanded && (
                  <div className="border-t border-border">
                    {children.map((cat) => (
                      <div
                        key={cat.id}
                        className="flex items-center gap-2 border-b border-border/50 px-3 py-2 last:border-b-0"
                      >
                        <span className="w-6 flex-shrink-0" />
                        <span
                          className="h-2 w-2 flex-shrink-0 rounded-full"
                          style={{ background: group.color ?? 'var(--muted-foreground)' }}
                        />
                        <span className="flex-1 text-sm">{cat.name}</span>
                        <Button
                          size="sm"
                          variant="ghost"
                          className="h-7 px-2 text-muted-foreground hover:text-foreground"
                          onClick={() => openEditCategory(cat)}
                        >
                          <Pencil className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    ))}
                    <div className="px-3 py-2">
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-7 text-xs text-muted-foreground"
                        onClick={() => openCreateCategory(group)}
                      >
                        <Plus className="h-3 w-3" />
                        Add category
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      <GroupDialog
        open={groupDialogOpen}
        onOpenChange={(open) => {
          setGroupDialogOpen(open)
          if (!open) setEditingGroup(null)
        }}
        group={editingGroup ?? undefined}
      />

      <CategoryItemDialog
        open={catDialogOpen}
        onOpenChange={(open) => {
          setCatDialogOpen(open)
          if (!open) {
            setCatDialogParentGroup(null)
            setEditingCategory(null)
          }
        }}
        category={editingCategory ?? undefined}
        parentGroup={catDialogParentGroup ?? undefined}
      />
    </div>
  )
}
