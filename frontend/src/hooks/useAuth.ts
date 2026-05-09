import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ApiError } from '@/api/client'
import * as authApi from '@/api/auth'
import type { AuthCredentials, UpdateUserRequest } from '@/types/api'
import type { User } from '@/types/models'

export const meQueryKey = ['me'] as const

export function useAuth() {
  const queryClient = useQueryClient()

  const meQuery = useQuery<User | null>({
    queryKey: meQueryKey,
    queryFn: async () => {
      try {
        return await authApi.getMe()
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) return null
        throw err
      }
    },
    retry: false,
    staleTime: 60_000,
  })

  const loginMutation = useMutation({
    mutationFn: (credentials: AuthCredentials) => authApi.login(credentials),
    onSuccess: (res) => {
      queryClient.setQueryData(meQueryKey, res.user)
    },
  })

  const registerMutation = useMutation({
    mutationFn: (credentials: AuthCredentials) => authApi.register(credentials),
    onSuccess: (res) => {
      queryClient.setQueryData(meQueryKey, res.user)
    },
  })

  const logoutMutation = useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      queryClient.setQueryData(meQueryKey, null)
      queryClient.clear()
    },
  })

  const updateMeMutation = useMutation({
    mutationFn: (body: UpdateUserRequest) => authApi.updateMe(body),
    onSuccess: (user) => {
      queryClient.setQueryData(meQueryKey, user)
    },
  })

  return {
    user: meQuery.data ?? null,
    isLoading: meQuery.isLoading,
    isAuthenticated: !!meQuery.data,
    refetchMe: meQuery.refetch,
    login: loginMutation,
    register: registerMutation,
    logout: logoutMutation,
    updateMe: updateMeMutation,
  }
}
