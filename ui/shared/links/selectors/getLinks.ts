import Link from 'models/Link';
import { AppState } from 'store';
import { Map } from 'immutable';

export const getLinks = (state: AppState): Map<number, Link> => state.links.items;
