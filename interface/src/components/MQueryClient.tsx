import React, { useCallback, useMemo } from 'react';
import { QueryClient, QueryClientProvider, QueryFunctionContext, QueryKey } from '@tanstack/react-query';
import { AxiosError, AxiosInstance } from 'axios';

import monetrClient from '@monetr/interface/api/api';

export enum QueryMethod {
  UseQuery,
  UseBody,
}

export interface MQueryClientProps {
  children: React.ReactElement;
  client?: AxiosInstance;
}

export default function MQueryClient(props: MQueryClientProps): JSX.Element {
  const client = useMemo(() => {
    if (props.client) {
      return client;
    }

    return monetrClient;
  }, [props.client]);

  const queryFn = useCallback(async (context: QueryFunctionContext<QueryKey>) => {
    let method = 'GET';
    let body = undefined;
    let params = {};
    if (context.queryKey.length > 1 && context?.meta?.['method'] !== QueryMethod.UseQuery) {
      method = 'POST';
      body = context.queryKey[1];
    }

    if (context?.meta?.['method'] === QueryMethod.UseQuery && context.queryKey.length > 1) {
      params = context.queryKey[1];
    }

    if (context.pageParam) {
      params['offset'] = context.pageParam;
    }

    console.warn({
      url: `${context.queryKey[0]}`,
      method: method,
      params: params,
      data: body,
    });

    const { data } = await client.request({
      url: `${context.queryKey[0]}`,
      method: method,
      params: params,
      data: body,
    })
      .catch((result: AxiosError) => {
        switch (result.response.status) {
          case 404:
          case 500: // Internal Server Error
          case 502:
            throw result;
          default:
            return result.response;
        }
      });
    return data;
  }, [client]);

  const queryClient = useMemo(() => new QueryClient({
    // TODO make this configurable somehow? Its annoying in tests but maybe good for local dev?
    logger: {
      log: () => { },
      warn: () => { },
      error: () => { },
    },
    defaultOptions: {
      queries: {
        staleTime: 10 * 60 * 1000, // 10 minute default stale time,
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
