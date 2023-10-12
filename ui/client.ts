import { QueryClient } from '@tanstack/react-query';

import Query from 'util/query';

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

export default queryClient;
