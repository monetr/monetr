import { Map } from 'immutable';
import Link from 'models/Link';
import { AppState } from 'store';

export const getLinks = (state: AppState): Map<number, Link> => state.links.items;
