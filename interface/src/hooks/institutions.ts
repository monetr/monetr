import { useQuery, UseQueryResult } from '@tanstack/react-query';

import Institution from 'models/Institution';

export type InstitutionResult =
  { result: Institution }
  & UseQueryResult<Partial<Institution>>;

export function useInstitution(institutionId: string | null): InstitutionResult {
  const result = useQuery<{ logo: string }>([`/institutions/${ institutionId }`], {
    staleTime: 30 * 60 * 1000, // 30 minutes
    enabled: Boolean(institutionId),
  });
  return {
    ...result,
    result: new Institution(result.data),
  };
}
