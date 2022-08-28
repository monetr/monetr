import { useQueryClient } from 'react-query';

import request from 'util/request';

export default function useLogout(): () => Promise<void> {
  const queryClient = useQueryClient();
  return async () => {
    return request()
      .get('/authentication/logout')
      .then(() => queryClient.invalidateQueries('/users/me'));
  };
}
