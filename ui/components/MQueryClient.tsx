import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { ReactElement } from './types';

import Query from 'util/query';

export interface MQueryClientProps {
  children: ReactElement;
}

export default function MQueryClient(props: MQueryClientProps): JSX.Element {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 10 * 60 * 1000, // 10 minute default stale time,
        queryFn: Query,
      },
    },
  });

  return (
    <QueryClientProvider client={ queryClient }>
      { props.children }
    </QueryClientProvider>
  );
}
