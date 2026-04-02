import { useQueryClient } from '@tanstack/react-query';

import request from '@monetr/interface/util/request';

export default function useLogout(): () => Promise<void> {
  const queryClient = useQueryClient();
  return async () => {
    return request({ method: 'GET', url: '/api/authentication/logout' }).then(() =>
      queryClient.invalidateQueries({ queryKey: ['/api/users/me'] }),
    );
  };
}
