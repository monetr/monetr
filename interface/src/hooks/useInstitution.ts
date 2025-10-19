import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import Institution from '@monetr/interface/models/Institution';

export function useInstitution(institutionId?: string): UseQueryResult<Institution, unknown> {
  return useQuery<Partial<Institution>, unknown, Institution>({
    queryKey: [`/institutions/${institutionId}`],
    staleTime: 30 * 60 * 1000, // 30 minutes
    enabled: Boolean(institutionId),
    select: data => new Institution(data),
  });
}
