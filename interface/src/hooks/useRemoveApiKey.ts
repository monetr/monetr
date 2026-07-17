import { useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';

import type ApiKey from '@monetr/interface/models/ApiKey';
import type { ID } from '@monetr/interface/models/ID';
import type { WithJsonValues } from '@monetr/interface/util/json';
import request from '@monetr/interface/util/request';

export type RemoveApiKeyRequest =
  | {
      apiKeyId: ID<ApiKey>;
      challenge: never;
      nonce: never;
    }
  | {
      apiKeyId: ID<ApiKey>;
      challenge: string;
      nonce: number;
    };

export default function useRemoveApiKey(): (_: RemoveApiKeyRequest) => Promise<void> {
  const removeApiKey = useCallback(async (data: RemoveApiKeyRequest): Promise<void> => {
    return void (await request({
      method: 'DELETE',
      url: `/api/keys/${data.apiKeyId}`,
      data:
        // Only send a body if there is a challenge and nonce
        data.challenge && data.nonce
          ? {
              challenge: data.challenge,
              nonce: data.nonce,
            }
          : undefined,
    }));
  }, []);

  const { mutateAsync } = useMutation({
    mutationFn: removeApiKey,
    onSuccess: (_, input, _result, { client: queryClient }) =>
      Promise.all([
        queryClient.setQueryData([`/api/keys`], (previous: Array<WithJsonValues<ApiKey>>) =>
          previous.filter(item => item.apiKeyId !== input.apiKeyId),
        ),
      ]),
  });

  return mutateAsync;
}
