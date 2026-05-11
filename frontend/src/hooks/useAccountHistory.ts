import { useQuery } from '@tanstack/react-query'
import * as accountsApi from '@/api/accounts'
import type { AccountHistoryQuery } from '@/types/api'

export function useAccountHistory(query: AccountHistoryQuery) {
  return useQuery({
    queryKey: ['accounts', 'history', query],
    queryFn: () => accountsApi.getAccountHistory(query),
    staleTime: 60_000,
  })
}
