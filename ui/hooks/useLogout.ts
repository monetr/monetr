import { useQueryClient } from 'react-query';
import { useDispatch } from 'react-redux';

import { Logout } from 'shared/authentication/actions';
import request from 'shared/util/request';

export default function useLogout(): () => Promise<void> {
  const dispatch = useDispatch();
  const queryClient = useQueryClient();
  return () => {
    dispatch({
      type: Logout.Pending,
    });

    return request()
      .get('/authentication/logout')
      .then(() => queryClient.invalidateQueries('/api/users/me'))
      .then(() => void dispatch({
        type: Logout.Success,
      }))
      .catch(error => {
        dispatch({
          type: Logout.Failure,
        });

        throw error;
      });
  };
}
