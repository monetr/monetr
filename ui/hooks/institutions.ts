import { useQuery, UseQueryResult } from 'react-query';

import Institution from 'models/Institution';

export type InstitutionResult =
  { result: Institution }
  & UseQueryResult<Partial<Institution>>;

export function useInstitution(institutionId: string): InstitutionResult {
  const result = useQuery<{ logo: string }>(`/institutions/${ institutionId }`, {
    staleTime: 30 * 60 * 1000, // 30 minutes
  });
  return {
    ...result,
    result: new Institution(result.data),
  };
}
