import type React from 'react';
import { useCallback, useMemo } from 'react';
import { QueryClient, QueryClientProvider, type QueryFunctionContext, type QueryKey } from '@tanstack/react-query';

import type { RequestConfig } from '@monetr/interface/api/client';
import request, { type ApiError } from '@monetr/interface/util/request';

export interface MQueryClientProps {
  children: React.ReactElement;
}

interface RequestParams {
  offset?: string | number;
  [key: string]: string | number | boolean | undefined;
}

export default function MQueryClient(props: MQueryClientProps): React.JSX.Element {
  const queryFn = useCallback(async (context: QueryFunctionContext<QueryKey>) => {
    // Query keys are always [method, path, params?]. For POST and QUERY the params are sent as the request body,
    // otherwise they are sent as query params.
    const [method, url, keyParams] = context.queryKey as [RequestConfig['method'], string, RequestParams | undefined];
    let body: unknown;
    let params: RequestParams = {};
    if (method === 'POST' || method === 'QUERY') {
      body = keyParams;
    } else {
      // Copy the params so that adding the offset below does not mutate the query key itself.
      params = { ...keyParams };
    }

    if (context.pageParam) {
      params.offset = context.pageParam as string | number;
    }

    const { data } = await request({
      method: method,
      url: url,
      params: params,
      data: body,
    }).catch((error: ApiError) => {
      switch (error.response.status) {
        case 404:
        case 500: // Internal Server Error
        case 502:
          throw error;
        default:
          return error.response;
      }
    });
    return data;
  }, []);

  const queryClient = useMemo(
    () =>
      new QueryClient({
        // TODO make this configurable somehow? Its annoying in tests but maybe good for local dev?
        defaultOptions: {
          queries: {
            staleTime: 10 * 60 * 1000, // 10 minute default stale time,
            queryFn: queryFn,
          },
        },
      }),
    [queryFn],
  );

  return <QueryClientProvider client={queryClient}>{props.children}</QueryClientProvider>;
}
