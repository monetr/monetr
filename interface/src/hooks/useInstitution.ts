import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import Institution from '@monetr/interface/models/Institution';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useInstitution(institutionId?: string): UseQueryResult<Institution, unknown> {
  return useQuery<Omit<WithJsonValues<Institution>, 'timestamp'>, unknown, Institution>({
    queryKey: [`/api/institutions/${institutionId}`],
    staleTime: 30 * 60 * 1000, // 30 minutes
    enabled: Boolean(institutionId),
    select: data => new Institution(data),
  });
}
