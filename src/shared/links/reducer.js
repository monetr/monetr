import LinksState from "./state";
import {FETCH_LINKS_FAILURE, FETCH_LINKS_REQUEST, FETCH_LINKS_SUCCESS} from "./actions";

export default function reducer(state = new LinksState(), action) {
  switch (action.type) {
    case FETCH_LINKS_REQUEST:
      return state.merge({
        loading: true,
      });
    case FETCH_LINKS_FAILURE:
      return state.merge({
        loading: false,
      });
    case FETCH_LINKS_SUCCESS:
      return state.merge({
        loaded: true,
        loading: false,
        links: action.payload,
      });
    default:
      return state;
  }
}
