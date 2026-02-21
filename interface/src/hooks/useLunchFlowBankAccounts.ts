import { type UseQueryResult, useQuery } from '@tanstack/react-query';

import LunchFlowBankAccount from '@monetr/interface/models/LunchFlowBankAccount';

export function useLunchFlowBankAccounts(
  lunchFlowLinkId?: string,
): UseQueryResult<Array<LunchFlowBankAccount>, unknown> {
  return useQuery<Array<Partial<LunchFlowBankAccount>>, unknown, Array<LunchFlowBankAccount>>({
    queryKey: [`/lunch_flow/link/${lunchFlowLinkId}/bank_accounts`],
    enabled: Boolean(lunchFlowLinkId),
    select: data => (data ?? []).map(item => new LunchFlowBankAccount(item)),
  });
}
