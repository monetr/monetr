import { getLinks } from "shared/links/selectors/getLinks";
import { Map } from 'immutable';
import { createSelector } from "reselect";
import Link from "models/Link";

export const getLink = (linkId: number) => createSelector<any, any, Link | null>(
  [getLinks],
  (links: Map<number, Link>) => links.get(linkId, null),
);
