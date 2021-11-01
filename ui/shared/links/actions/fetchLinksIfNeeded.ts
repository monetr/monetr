import { getHasAnyLinks } from 'shared/links/selectors/getHasAnyLinks';
import fetchLinks from 'shared/links/actions/fetchLinks';
import { AppActionWithState, AppDispatch, GetAppState } from 'store';

export default function fetchLinksIfNeeded(): AppActionWithState<Promise<void>> {
  return (dispatch: AppDispatch, getState: GetAppState): Promise<void> => {
    const hasLinks = getHasAnyLinks(getState());
    if (hasLinks) {
      return Promise.resolve();
    }

    return fetchLinks()(dispatch);
  }
}
