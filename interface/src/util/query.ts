import { QueryFunctionContext, QueryKey } from '@tanstack/query-core';
import axios from 'axios';

export default async function Query<T = unknown, TQueryKey extends QueryKey = QueryKey>(
  context: QueryFunctionContext<TQueryKey>,
): Promise<T> {
  const { data } = await axios.request<T>({
    url: `/api${context.queryKey[0]}`,
    method: context.queryKey.length === 1 ? 'GET' : 'POST',
    params: context.pageParam && {
      offset: context.pageParam,
    },
    data: context.queryKey.length === 2 && context.queryKey[1],
  })
    .catch(result => {
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
}
