import { useStore } from 'react-redux';
import fetchLinks from 'shared/links/actions/fetchLinks';
import { getHasAnyLinks } from 'shared/links/selectors/getHasAnyLinks';

export default function useFetchLinksIfNeeded(): () => Promise<void> {
  const { dispatch, getState } = useStore();

  return function (): Promise<void> {
    const hasLinks = getHasAnyLinks(getState());
    if (hasLinks) {
      return Promise.resolve();
    }

    return fetchLinks()(dispatch);
  };
}
