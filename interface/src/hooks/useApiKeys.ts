import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import ApiKey from '@monetr/interface/models/ApiKey';
import type { WithJsonValues } from '@monetr/interface/util/json';

export default function useApiKeys(): UseQueryResult<Array<ApiKey>, unknown> {
  return useQuery<Array<WithJsonValues<ApiKey>>, unknown, Array<ApiKey>>({
    queryKey: [`/api/keys`],
    select: data => (data || []).map(item => new ApiKey(item)),
  });
}
