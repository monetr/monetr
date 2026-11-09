import { useQueryClient } from '@tanstack/react-query';

import request from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

export function useTriggerManualSimpleFinSync(): (_linkId: string) => Promise<void> {
  const { enqueueSnackbar } = useSnackbar();
  const queryClient = useQueryClient();
  return async (linkId: string): Promise<void> => {
    return await request({
      method: 'POST',
      url: '/api/simplefin/link/sync',
      data: {
        linkId,
      },
    })
      .then(
        () =>
          void enqueueSnackbar('Triggered a manual sync in the background!', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
      )
      // Will make things like the "last attempted update" timestamp thing update. The status card reads the single link
      // which is a different key from the list, so both need to be invalidated.
      .then(
        () =>
          void setTimeout(() => {
            queryClient.invalidateQueries({ queryKey: ['GET', '/api/links'] });
            queryClient.invalidateQueries({ queryKey: ['GET', `/api/links/${linkId}`] });
          }, 2000),
      )
      .catch(
        error =>
          void enqueueSnackbar(`Failed to trigger a manual sync: ${error?.response?.data?.error || 'unknown error'}.`, {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      );
  };
}
