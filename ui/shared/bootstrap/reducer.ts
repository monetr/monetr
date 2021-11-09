import BootstrapState from 'shared/bootstrap/state';
import { Bootstrap, BootstrapActions } from 'shared/bootstrap/actions';

export default function reducer(state = new BootstrapState(), action: BootstrapActions): BootstrapState {
  switch (action.type) {
    case Bootstrap.Begin:
      return {
        ...state,
        isReady: false,
        isBootstrapping: true,
      };
    case Bootstrap.Failure:
      return {
        ...state,
        isReady: false,
        isBootstrapping: false,
      };
    case Bootstrap.Success:
      return {
        ...state,
        ...action.payload,
        isReady: true,
        isBootstrapping: false,
      }
    default:
      return state;
  }
}
