import { useStore } from 'react-redux';
import { getHasAnyLinks } from 'shared/links/selectors/getHasAnyLinks';
import fetchLinks from 'shared/links/actions/fetchLinks';

export default function useFetchLinksIfNeeded(): () => Promise<void> {
  const { dispatch, getState } = useStore();

  return function (): Promise<void> {
    const hasLinks = getHasAnyLinks(getState());
    if (hasLinks) {
      return Promise.resolve();
    }

    return fetchLinks()(dispatch);
  }
}
