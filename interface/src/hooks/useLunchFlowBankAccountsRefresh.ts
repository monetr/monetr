import { type UseMutationResult, useMutation } from '@tanstack/react-query';
import { HttpStatusCode } from 'axios';

import request from '@monetr/interface/util/request';

export default function useLunchFlowBankAccountsRefresh(): UseMutationResult<string, Error, string, unknown> {
  return useMutation({
    mutationFn: async (lunchFlowLinkId?: string): Promise<string> => {
      return request()
        .post(`/lunch_flow/link/${lunchFlowLinkId}/bank_accounts/refresh`)
        .then(result => {
          if (result.status !== HttpStatusCode.NoContent) {
            throw result;
          }
        })
        .then(() => lunchFlowLinkId);
    },
    onSuccess: (lunchFlowLinkId: string, _a, _b, context) =>
      context.client.invalidateQueries({ queryKey: [`/lunch_flow/link/${lunchFlowLinkId}/bank_accounts`] }),
  });
}
