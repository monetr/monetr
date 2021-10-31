import { AppState } from 'store';

export const getHasAnyLinks = (state: AppState): boolean => state.links.items.count() > 0;
