import type { Category } from '@/types/models'

export function resolveGroupId(categoryId: string | null, catMap: Map<string, Category>): string {
  if (!categoryId) return '__uncategorized'
  const cat = catMap.get(categoryId)
  if (!cat) return '__uncategorized'
  return cat.parent_id ?? cat.id
}
