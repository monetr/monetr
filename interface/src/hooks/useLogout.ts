import { useQueryClient } from '@tanstack/react-query';

import request from '@monetr/interface/util/request';

export default function useLogout(): () => Promise<void> {
  const queryClient = useQueryClient();
  return async () => {
    return request()
      .get('/authentication/logout')
      .then(() => queryClient.invalidateQueries(['/users/me']));
  };
}
