import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as importsApi from '@/api/imports'
import { accountsKey } from './useAccounts'

export const importsKey = ['imports'] as const
export const importJobKey = (id: string) => ['imports', id] as const

export function useImports() {
  return useQuery({
    queryKey: importsKey,
    queryFn: () => importsApi.listImports(),
    refetchInterval: (q) => {
      const data = q.state.data
      if (!data) return false
      const hasInflight = data.some((j) => j.status === 'pending' || j.status === 'processing')
      return hasInflight ? 2000 : false
    },
  })
}

export function useImportJob(id: string | null) {
  return useQuery({
    queryKey: importJobKey(id ?? '__none__'),
    queryFn: () => importsApi.getImportJob(id!),
    enabled: !!id,
    refetchInterval: (q) => {
      const data = q.state.data
      if (!data) return 2000
      return data.status === 'pending' || data.status === 'processing' ? 2000 : false
    },
  })
}

export function useUploadCsv() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ file, accountId }: { file: File; accountId: string }) =>
      importsApi.uploadCsv(file, accountId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: importsKey })
      queryClient.invalidateQueries({ queryKey: ['transactions'] })
      queryClient.invalidateQueries({ queryKey: accountsKey })
    },
  })
}
