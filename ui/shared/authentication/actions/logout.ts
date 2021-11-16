import { useDispatch } from 'react-redux';
import { Logout } from 'shared/authentication/actions';
import request from 'shared/util/request';

export default function useLogout(): () => Promise<void> {
  const dispatch = useDispatch();
  return () => {
    dispatch({
      type: Logout.Pending,
    });

    return request()
      .get('/authentication/logout')
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
