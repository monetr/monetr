import LinksState from "shared/links/state";
import { FETCH_LINKS_FAILURE, FETCH_LINKS_REQUEST, FETCH_LINKS_SUCCESS, LinkActions } from "shared/links/actions";
import { LOGOUT } from "shared/authentication/actions";
import Link from "data/Link";

export default function reducer(state: LinksState = new LinksState(), action: LinkActions) {
  switch (action.type) {
    case FETCH_LINKS_REQUEST:
      return {
        ...state,
        loading: true,
      };
    case FETCH_LINKS_FAILURE:
      return {
        ...state,
        loading: false,
      };
    case FETCH_LINKS_SUCCESS:
      return {
        ...state,
        loaded: true,
        loading: false,
        items: state.items.merge(action.payload),
      };
    case LOGOUT:
      return new LinksState();
    default:
      return state;
  }
}
