import { ACTIVATE_SUBSCRIPTION } from 'shared/authentication/actions';

export default function activateSubscription() {
  return dispatch => {
    return dispatch({
      type: ACTIVATE_SUBSCRIPTION,
    });
  };
}
