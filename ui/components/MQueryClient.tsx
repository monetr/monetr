import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { ReactElement } from './types';

import Query from 'util/query';

export interface MQueryClientProps {
  children: ReactElement;
}

export default function MQueryClient(props: MQueryClientProps): JSX.Element {
  const queryClient = new QueryClient({
    // TODO make this configurable somehow? Its annoying in tests but maybe good for local dev?
    logger: {
      log: () => { },
      warn: () => { },
      error: () => { },
    },
    defaultOptions: {
      queries: {
        staleTime: 10 * 60 * 1000, // 10 minute default stale time,
        queryFn: Query,
      },
    },
  });

  return (
    <QueryClientProvider client={ queryClient }>
      {props.children}
    </QueryClientProvider>
  );
}
