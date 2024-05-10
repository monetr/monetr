import { useQueryClient } from '@tanstack/react-query';
import axios from 'axios';

export default function useLogout(): () => Promise<void> {
  const queryClient = useQueryClient();
  return async () => {
    return axios
      .get('/api/authentication/logout')
      .then(() => queryClient.invalidateQueries(['/users/me']));
  };
}
