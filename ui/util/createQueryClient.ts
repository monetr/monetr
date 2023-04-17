import axios from "axios";
import { QueryClient, QueryFunctionContext, QueryKey } from "react-query";

async function queryFn<T = unknown, TQueryKey extends QueryKey = QueryKey>(
  context: QueryFunctionContext<TQueryKey>,
): Promise<T> {
  const { data } = await axios.request<T>({
    url: `/api${ context.queryKey[0] }`,
    method: context.queryKey.length === 1 ? 'GET' : 'POST',
    params: context.pageParam && {
      offset: context.pageParam,
    },
    data: context.queryKey.length === 2 && context.queryKey[1],
  })
    .catch(result => {
      switch (result.response.status) {
        case 500: // Internal Server Error
          throw result;
        default:
          return result.response;
      }
    });
  return data;
}

function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 10 * 60 * 1000, // 10 minute default stale time,
        queryFn: queryFn,
      },
    },
  })
}

export default createQueryClient;
