import {getHasAnyLinks} from "../selectors/getHasAnyLinks";
import fetchLinks from "./fetchLinks";

export default function fetchLinksIfNeeded() {
  return (dispatch, getState) => {
    const hasLinks = getHasAnyLinks(getState());
    if (hasLinks) {
      return Promise.resolve();
    }

    return dispatch(fetchLinks());
  }
}
