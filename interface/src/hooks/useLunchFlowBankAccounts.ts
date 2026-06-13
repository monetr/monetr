import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import LunchFlowBankAccount from '@monetr/interface/models/LunchFlowBankAccount';
import type { WithJsonValues } from '@monetr/interface/util/json';

export function useLunchFlowBankAccounts(
  lunchFlowLinkId?: string,
): UseQueryResult<Array<LunchFlowBankAccount>, unknown> {
  return useQuery<Array<WithJsonValues<LunchFlowBankAccount>>, unknown, Array<LunchFlowBankAccount>>({
    queryKey: [`/api/lunch_flow/link/${lunchFlowLinkId}/bank_accounts`],
    enabled: Boolean(lunchFlowLinkId),
    select: data => (data ?? []).map(item => new LunchFlowBankAccount(item)),
  });
}
