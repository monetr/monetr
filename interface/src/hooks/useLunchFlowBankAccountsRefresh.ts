import { type UseMutationResult, useMutation } from '@tanstack/react-query';

import type { ApiError } from '@monetr/interface/api/client';
import request from '@monetr/interface/util/request';

export default function useLunchFlowBankAccountsRefresh(): UseMutationResult<
  string,
  ApiError<{ error: string }>,
  string,
  unknown
> {
  return useMutation({
    mutationFn: async (lunchFlowLinkId?: string): Promise<string> => {
      return request({ method: 'POST', url: `/api/lunch_flow/link/${lunchFlowLinkId}/bank_accounts/refresh` })
        .then(result => {
          if (result.status !== 204) {
            throw result;
          }
        })
        .then(() => lunchFlowLinkId);
    },
    onSuccess: (lunchFlowLinkId: string, _a, _b, context) =>
      context.client.invalidateQueries({ queryKey: [`/api/lunch_flow/link/${lunchFlowLinkId}/bank_accounts`] }),
  });
}
