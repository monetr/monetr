import { AxiosError } from 'axios';
import User from 'models/User';
import { useDispatch } from 'react-redux';
import request from 'shared/util/request';
import { Login } from 'shared/authentication/actions';

export default function useBootstrapLogin(): (user?: User | null, subscriptionIsActive?: boolean) => Promise<void> {
  const dispatch = useDispatch();
  return (user?: User | null, subscriptionIsActive: boolean = true) => {
    dispatch({
      type: Login.Pending,
    });

    // If a user object is provided to this method then that means it was received from something like the login
    // endpoint. This saves us a network round trip if the user just logged in.
    if (user) {
      dispatch({
        type: Login.Success,
        payload: {
          user: user,
          isActive: subscriptionIsActive,
        }
      });
      return Promise.resolve();
    }

    return request().get('/users/me')
      .then(result => void dispatch({
        type: Login.Success,
        payload: {
          user: new User(result.data.user),
          isActive: result.data.isActive,
        },
      }))
      .catch((error: AxiosError) => {
        dispatch({
          type: Login.Failure,
        });

        // If we are not authenticated then don't through. This is going to be acceptable behavior.
        if (error.response.status === 403) {
          return;
        }

        // Any other scenarios should throw the exception though.
        throw error;
      });
  }
}
