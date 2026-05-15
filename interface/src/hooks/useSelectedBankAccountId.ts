import { useMemo } from 'react';
import { useRoute } from 'wouter';

export function useSelectedBankAccountId(): string | undefined {
  const [, params] = useRoute<{ bankId: string }>('/bank/:bankId/*');
  return useMemo(() => {
    return params?.bankId || undefined;
  }, [params]);
}
