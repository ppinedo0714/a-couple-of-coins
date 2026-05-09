import { apiFetch } from './client'
import type { AuthCredentials, AuthResponse, UpdateUserRequest } from '@/types/api'
import type { User } from '@/types/models'

export function login(credentials: AuthCredentials) {
  return apiFetch<AuthResponse>('/auth/login', { method: 'POST', body: credentials })
}

export function register(credentials: AuthCredentials) {
  return apiFetch<AuthResponse>('/auth/register', { method: 'POST', body: credentials })
}

export function logout() {
  return apiFetch<void>('/auth/logout', { method: 'POST' })
}

export function getMe() {
  return apiFetch<User>('/users/me')
}

export function updateMe(body: UpdateUserRequest) {
  return apiFetch<User>('/users/me', { method: 'PUT', body })
}
