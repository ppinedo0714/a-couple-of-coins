import { apiFetch } from './client'
import type { Category } from '@/types/models'
import type { CreateCategoryRequest, UpdateCategoryRequest } from '@/types/api'

export function listCategories() {
  return apiFetch<Category[]>('/categories')
}

export function createCategory(body: CreateCategoryRequest) {
  return apiFetch<Category>('/categories', { method: 'POST', body })
}

export function updateCategory(id: string, body: UpdateCategoryRequest) {
  return apiFetch<Category>(`/categories/${id}`, { method: 'PUT', body })
}

export function deleteCategory(id: string) {
  return apiFetch<void>(`/categories/${id}`, { method: 'DELETE' })
}
