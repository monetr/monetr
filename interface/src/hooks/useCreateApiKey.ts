import { useCallback } from 'react';
import { useMutation } from '@tanstack/react-query';

import ApiKey from '@monetr/interface/models/ApiKey';
import type { WithJsonValues } from '@monetr/interface/util/json';
import type { Writable } from '@monetr/interface/util/readonly';
import request from '@monetr/interface/util/request';

export type CreateApiKeyRequest = Writable<ApiKey> & {
  challenge?: string;
  nonce?: number;
};

export type CreateApiKeyResponse = ApiKey & {
  secret: string;
};

export default function useCreateApiKey(): (_: CreateApiKeyRequest) => Promise<CreateApiKeyResponse> {
  const createApiKey = useCallback(async (data: CreateApiKeyRequest): Promise<CreateApiKeyResponse> => {
    return await request<WithJsonValues<CreateApiKeyResponse>>({
      method: 'POST',
      url: `/api/keys`,
      data,
    }).then(result => ({
      ...new ApiKey(result.data),
      secret: result.data.secret,
    }));
  }, []);

  const { mutateAsync } = useMutation({
    mutationFn: createApiKey,
    onSuccess: (_data, _var, _result, { client: queryClient }) =>
      Promise.all([queryClient.invalidateQueries({ queryKey: [`/api/keys`] })]),
  });

  return mutateAsync;
}
