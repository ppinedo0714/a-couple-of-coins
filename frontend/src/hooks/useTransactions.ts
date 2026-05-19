import { useMutation, useQuery, useQueryClient, type UseQueryOptions } from '@tanstack/react-query'
import * as transactionsApi from '@/api/transactions'
import { accountsKey } from './useAccounts'
import type {
  CreateTransactionRequest,
  ListTransactionsQuery,
  ListTransactionsResponse,
  UpdateTransactionRequest,
} from '@/types/api'

export const transactionsKey = (query: ListTransactionsQuery = {}) => ['transactions', query] as const

export function useTransactions(
  query: ListTransactionsQuery = {},
  options?: Omit<UseQueryOptions<ListTransactionsResponse>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: transactionsKey(query),
    queryFn: () => transactionsApi.listTransactions(query),
    ...options,
  })
}

export function useCreateTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateTransactionRequest) => transactionsApi.createTransaction(body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: accountsKey })
    },
  })
}

export function useUpdateTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateTransactionRequest }) =>
      transactionsApi.updateTransaction(id, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: accountsKey })
    },
  })
}

export function useDeleteTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => transactionsApi.deleteTransaction(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: accountsKey })
    },
  })
}

export function useClassifyTransactions() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => transactionsApi.classifyTransactions(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['transactions'] }),
  })
}
