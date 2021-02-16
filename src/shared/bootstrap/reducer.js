import BootstrapState from "shared/bootstrap/state";
import { BOOTSTRAP_FAILED, BOOTSTRAP_FINISHED, BOOTSTRAP_START } from "shared/bootstrap/actions";

export default function reducer(state = new BootstrapState(), action) {
  switch (action.type) {
    case BOOTSTRAP_START:
      return state.merge({
        isReady: false,
        isBootstrapping: true,
      });
    case BOOTSTRAP_FAILED:
      return state.merge({
        isReady: false,
        isBootstrapping: false,
      });
    case BOOTSTRAP_FINISHED:
      return state.merge({
        isReady: true,
        isBootstrapping: false,
        ...action.config,
      });
    default:
      return state;
  }
}
