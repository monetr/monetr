import { useMemo } from 'react';
import { useMatch } from 'react-router-dom';

export function useSelectedBankAccountId(): string | undefined {
  const match = useMatch('/bank/:bankId/*');
  return useMemo(() => {
    return match?.params?.bankId || undefined;
  }, [match]);
}
