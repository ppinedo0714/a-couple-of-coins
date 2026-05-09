import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as accountsApi from '@/api/accounts'
import type { CreateAccountRequest, UpdateAccountRequest } from '@/types/api'

export const accountsKey = ['accounts'] as const

export function useAccounts() {
  return useQuery({
    queryKey: accountsKey,
    queryFn: () => accountsApi.listAccounts(),
  })
}

export function useCreateAccount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateAccountRequest) => accountsApi.createAccount(body),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: accountsKey }),
  })
}

export function useUpdateAccount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateAccountRequest }) =>
      accountsApi.updateAccount(id, body),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: accountsKey }),
  })
}

export function useDeleteAccount() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => accountsApi.deleteAccount(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: accountsKey }),
  })
}
