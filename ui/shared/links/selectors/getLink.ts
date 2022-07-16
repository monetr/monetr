import { Map } from 'immutable';
import Link from 'models/Link';
import { createSelector } from 'reselect';
import { getLinks } from 'shared/links/selectors/getLinks';

export const getLink = (linkId: number) => createSelector<any, any, Link | null>(
  [getLinks],
  (links: Map<number, Link>) => links.get(linkId, null),
);
