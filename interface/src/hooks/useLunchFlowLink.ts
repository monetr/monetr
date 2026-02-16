import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import LunchFlowLink from '@monetr/interface/models/LunchFlowLink';

export function useLunchFlowLink(lunchFlowLinkId?: string): UseQueryResult<LunchFlowLink, unknown> {
  return useQuery<Partial<LunchFlowLink>, unknown, LunchFlowLink>({
    queryKey: [`/lunch_flow/link/${lunchFlowLinkId}`],
    enabled: Boolean(lunchFlowLinkId),
    select: data => new LunchFlowLink(data),
  });
}
