import { type UseMutationResult, useMutation } from '@tanstack/react-query';

import type { ApiError } from '@monetr/interface/api/client';
import request from '@monetr/interface/util/request';

export default function useSimpleFinBankAccountsRefresh(): UseMutationResult<
  string,
  ApiError<{ error: string }>,
  string,
  unknown
> {
  return useMutation({
    mutationFn: async (simpleFinLinkId: string): Promise<string> => {
      return await request({ method: 'POST', url: `/api/simplefin/link/${simpleFinLinkId}/bank_accounts/refresh` })
        .then(result => {
          if (result.status !== 204) {
            throw result;
          }
        })
        .then(() => simpleFinLinkId);
    },
    onSuccess: (simpleFinLinkId: string, _a, _b, context) =>
      context.client.invalidateQueries({ queryKey: ['GET', `/api/simplefin/link/${simpleFinLinkId}/bank_accounts`] }),
  });
}
