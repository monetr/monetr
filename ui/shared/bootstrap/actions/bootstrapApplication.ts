import { NewClient } from 'api/api';
import { Bootstrap, BootstrapBegin } from 'shared/bootstrap/actions';
import BootstrapState from 'shared/bootstrap/state';
import request from 'shared/util/request';
import { AppAction, AppDispatch } from 'store';

const bootstrapBegin: BootstrapBegin = {
  type: Bootstrap.Begin,
};

export default function bootstrapApplication(): AppAction<Promise<void>> {
  return (dispatch: AppDispatch): Promise<void> => {
    dispatch(bootstrapBegin);

    window.API = NewClient({
      baseURL: '/api',
      withCredentials: true,
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