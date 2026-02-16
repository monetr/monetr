import { useEffect } from 'react';
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query';

export interface LunchFlowLinkSyncProgress {
  bankAccountId: string;
  status: 'begin' | 'transactions' | 'balances' | 'complete' | 'error';
}

export default function useLunchFlowLinkSyncProgress(
  linkId: string,
  bankAccountId: string,
): UseQueryResult<LunchFlowLinkSyncProgress, unknown> {
  const queryClient = useQueryClient();
  // Bootstrap the socket to listen for the actual changes.
  useEffect(() => {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
    const socket = new WebSocket(
      `${protocol}://${location.host}/api/lunch_flow/link/sync/${linkId}/bank_account/${bankAccountId}/progress`,
    );
    socket.onopen = () => {
      console.log('Listening for Lunch Flow sync progress', {
        linkId,
        bankAccountId,
      });
    };

    // Whenever we receive a progress message, update our state to represent the new status.
    socket.onmessage = event => {
      if (!event.data) {
        return;
      }
      const data: LunchFlowLinkSyncProgress = JSON.parse(event.data);
      const queryKey = [`/lunch_flow/link/sync/${linkId}/bank_account/${bankAccountId}/progress`];
      queryClient.setQueryData(queryKey, () => data);
    };

    // On unmount close the socket
    return () => socket.close();
  }, [linkId, bankAccountId, queryClient]);

  return useQuery<LunchFlowLinkSyncProgress, unknown, LunchFlowLinkSyncProgress>({
    queryKey: [`/lunch_flow/link/sync/${linkId}/bank_account/${bankAccountId}/progress`],
    initialData: () => null, // Don't do the initial fetch, rely on the websocket instead.
    staleTime: Infinity,
  });
}
