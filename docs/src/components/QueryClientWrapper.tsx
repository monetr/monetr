import { useCallback, useMemo } from 'react';
import { QueryClient, QueryClientProvider, QueryFunctionContext, QueryKey } from '@tanstack/react-query';

interface QueryClientWrapperProps {
  children: React.ReactNode;
}

export default function QueryClientWrapper(props: QueryClientWrapperProps): JSX.Element {
  const queryFn = useCallback(async (context: QueryFunctionContext<QueryKey>) => {
    const [url] = context.queryKey;
    const result = await fetch(url as string);
    const data = await result.json();
    return data;
  }, []);

  const queryClient = useMemo(() => new QueryClient({
    logger: {
      log: () => { },
      warn: () => { },
      error: () => { },
    },
    defaultOptions: {
      queries: {
        staleTime: 60 * 60 * 1000, // 60 minute default stale time,
        queryFn: queryFn,
      },
    },
  }), [queryFn]);

  return (
    <QueryClientProvider client={ queryClient }>
      { props.children }
    </QueryClientProvider>
  );
}
