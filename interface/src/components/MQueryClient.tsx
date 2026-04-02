import type React from 'react';
import { useCallback, useMemo } from 'react';
import { QueryClient, QueryClientProvider, type QueryFunctionContext, type QueryKey } from '@tanstack/react-query';

import request, { type ApiError } from '@monetr/interface/util/request';

export enum QueryMethod {
  UseQuery,
  UseBody,
}

export interface MQueryClientProps {
  children: React.ReactElement;
}

interface RequestParams {
  offset?: string | number;
  [key: string]: string | number | boolean | undefined;
}

export default function MQueryClient(props: MQueryClientProps): JSX.Element {
  const queryFn = useCallback(async (context: QueryFunctionContext<QueryKey>) => {
    let method: 'GET' | 'POST' = 'GET';
    let body: unknown;
    let params: RequestParams = {};
    if (context.queryKey.length > 1 && context?.meta?.method !== QueryMethod.UseQuery) {
      method = 'POST';
      body = context.queryKey[1];
    }

    if (context?.meta?.method === QueryMethod.UseQuery && context.queryKey.length > 1) {
      params = context.queryKey[1] as RequestParams;
    }

    if (context.pageParam) {
      params.offset = context.pageParam as string | number;
    }

    const { data } = await request({
      method: method,
      url: `${context.queryKey[0]}`,
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
