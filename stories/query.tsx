import React from 'react';
import {
  QueryClient, QueryClientProvider, QueryFunctionContext, QueryKey,
} from '@tanstack/react-query';

export interface MockRequest {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
  path: string;
  status: number;
  response: Object | Array<unknown>;
  once?: boolean;
}

export interface MockQueryClientProps {
  requests: Array<MockRequest>;
  children: React.ReactNode | JSX.Element;
}

export default function MockQueryClient(props: MockQueryClientProps): JSX.Element {
  let { requests } = props;
  async function queryFn<T = unknown, TQueryKey extends QueryKey = QueryKey>(
    context: QueryFunctionContext<TQueryKey>,
  ): Promise<T> {
    if (!requests || requests.length === 0) {
      return Promise.resolve({} as T);
    }
    const request = {
      url: `/api${ context.queryKey[0] }`,
      method: context.queryKey.length === 1 ? 'GET' : 'POST',
      params: context.pageParam && {
        offset: context.pageParam,
      },
      data: context.queryKey.length === 2 && context.queryKey[1],
    };
    const index = requests.findIndex(item => {
      // TODO Implement params matching.
      return item.path === request.url && item.method === request.method;
    });
    const response = requests[index];
    if (!response) {
      console.warn(`No response found for: ${ request.method } ${ request.url }`);
      return Promise.reject<T>({
        error: 'No response found!',
      });
    }

    // If the request is intended to only be handled a single time then remove the request from the stack.
    if (response.once) {
      requests = requests.splice(index, 1);
    }

    // Add a tiny bit of latency, as a treat.
    return new Promise<T>(resolve => {
      setTimeout(() => {
        resolve(response.response as T);
      }, 50);
    });
  }

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 10 * 60 * 1000, // 10 minute default stale time,
        queryFn: queryFn,
        retry: false,
      },
    },
  });

  return (
    <QueryClientProvider client={ queryClient } children={ props.children } />
  );
}
