import { getHasAnyLinks } from "shared/links/selectors/getHasAnyLinks";
import fetchLinks from "shared/links/actions/fetchLinks";

export default function fetchLinksIfNeeded() {
  return (dispatch, getState) => {
    const hasLinks = getHasAnyLinks(getState());
    if (hasLinks) {
      return Promise.resolve();
    }

    return dispatch(fetchLinks());
  }
}
