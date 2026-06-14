import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import LunchFlowLink from '@monetr/interface/models/LunchFlowLink';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useLunchFlowLink(lunchFlowLinkId?: string): UseQueryResult<LunchFlowLink, unknown> {
  return useQuery<WithJsonValues<LunchFlowLink>, unknown, LunchFlowLink>({
    queryKey: [`/api/lunch_flow/link/${lunchFlowLinkId}`],
    enabled: Boolean(lunchFlowLinkId),
    select: data => new LunchFlowLink(data),
  });
}
