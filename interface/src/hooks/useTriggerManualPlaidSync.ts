import { useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';

import request from '@monetr/interface/util/request';

export function useTriggerManualPlaidSync(): (_linkId: string) => Promise<void> {
  const { enqueueSnackbar } = useSnackbar();
  const queryClient = useQueryClient();
  return async (linkId: string): Promise<void> => {
    return request()
      .post('/plaid/link/sync', {
        linkId,
      })
      .then(() => void enqueueSnackbar('Triggered a manual sync in the background!', {
        variant: 'success',
        disableWindowBlurListener: true,
      }))
      // Will make things like the "last attempted update" timestamp thing update.
      .then(() => setTimeout(() => queryClient.invalidateQueries({ queryKey: ['/links'] }), 2000))
      .catch(error => void enqueueSnackbar(
        `Failed to trigger a manual sync: ${error?.response?.data?.error || 'unknown error'}.`,
        {
          variant: 'error',
          disableWindowBlurListener: true,
        },
      ));
  };
}
