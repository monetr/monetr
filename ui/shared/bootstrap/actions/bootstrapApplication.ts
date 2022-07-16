import { useDispatch } from 'react-redux';

import { NewClient } from 'api/api';
import { Bootstrap, BootstrapBegin } from 'shared/bootstrap/actions';
import BootstrapState from 'shared/bootstrap/state';
import request from 'shared/util/request';

const bootstrapBegin: BootstrapBegin = {
  type: Bootstrap.Begin,
};

export default function useBootstrapApplication(): () => Promise<void> {
  const dispatch = useDispatch();

  return (): Promise<void> => {
    dispatch(bootstrapBegin);

    window.API = NewClient({
      baseURL: '/api',
    });

    return request()
      .get('/config')
      .then(apiConfig => {
        dispatch({
          type: Bootstrap.Success,
          payload: new BootstrapState(apiConfig.data),
        });
      })
      .catch(error => {
        dispatch({
          type: Bootstrap.Failure,
        });

        throw error;
      });
  };
}
