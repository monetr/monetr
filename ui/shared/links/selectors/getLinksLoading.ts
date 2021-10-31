import { AppState } from 'store';

export const getLinksLoading = (state: AppState): boolean => state.links.loading;
