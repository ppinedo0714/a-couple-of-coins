import { useMemo } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as categoriesApi from '@/api/categories'
import type { CreateCategoryRequest, UpdateCategoryRequest } from '@/types/api'
import type { Category } from '@/types/models'

export const categoriesKey = ['categories'] as const

export function useCategories() {
  return useQuery({
    queryKey: categoriesKey,
    queryFn: () => categoriesApi.listCategories(),
  })
}

export function useGroupedCategories() {
  const { data: categories = [] } = useCategories()
  return useMemo(() => {
    const groups = categories.filter((c) => c.parent_id === null)
    const categoriesByGroupId: Record<string, Category[]> = {}
    for (const c of categories) {
      if (c.parent_id !== null) {
        if (!categoriesByGroupId[c.parent_id]) categoriesByGroupId[c.parent_id] = []
        categoriesByGroupId[c.parent_id]!.push(c)
      }
    }
    return { groups, categoriesByGroupId }
  }, [categories])
}

export function useCreateCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateCategoryRequest) => categoriesApi.createCategory(body),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: categoriesKey }),
  })
}

export function useUpdateCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateCategoryRequest }) =>
      categoriesApi.updateCategory(id, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: categoriesKey })
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
    },
  })
}

export function useDeleteCategory() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => categoriesApi.deleteCategory(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: categoriesKey })
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
    },
  })
}
