import { useEffect } from 'react';
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query';

export interface SimpleFinLinkSyncProgress {
  linkId: string;
  status: 'begin' | 'transactions' | 'balances' | 'complete' | 'error';
}

export default function useSimpleFinLinkSyncProgress(
  linkId: string,
): UseQueryResult<SimpleFinLinkSyncProgress | null, unknown> {
  const queryClient = useQueryClient();
  // Bootstrap the socket to listen for the actual changes.
  useEffect(() => {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
    const socket = new WebSocket(`${protocol}://${location.host}/api/simplefin/link/sync/${linkId}/progress`);
    socket.onopen = () => {
      console.log('Listening for SimpleFIN sync progress', {
        linkId,
      });
    };

    // Whenever we receive a progress message, update our state to represent the new status.
    socket.onmessage = event => {
      const data: SimpleFinLinkSyncProgress = JSON.parse(event.data);
      const queryKey = ['GET', `/api/simplefin/link/sync/${linkId}/progress`];
      queryClient.setQueryData(queryKey, data);
    };

    // On unmount close the socket
    return () => socket.close();
  }, [linkId, queryClient]);

  return useQuery<SimpleFinLinkSyncProgress | null, unknown, SimpleFinLinkSyncProgress | null>({
    queryKey: ['GET', `/api/simplefin/link/sync/${linkId}/progress`],
    initialData: () => null, // Don't do the initial fetch, rely on the websocket instead.
    staleTime: Infinity,
  });
}
