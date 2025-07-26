import { useQuery, UseQueryResult } from '@tanstack/react-query';

import Institution from '@monetr/interface/models/Institution';

export function useInstitution(institutionId: string | null): UseQueryResult<Institution> {
  return useQuery<Partial<Institution>, unknown, Institution>(
    [`/institutions/${ institutionId }`],
    {
      staleTime: 30 * 60 * 1000, // 30 minutes
      enabled: Boolean(institutionId),
      select: data => new Institution(data),
    },
  );
}
