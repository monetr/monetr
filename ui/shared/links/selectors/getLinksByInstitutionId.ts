import { Map } from 'immutable';
import Link from 'models/Link';
import { createSelector } from 'reselect';
import { getLinks } from 'shared/links/selectors/getLinks';

export const getLinksByInstitutionId = createSelector<any, any, Map<string, Link[]>>(
  [getLinks],
  (links: Map<number, Link>): Map<string, Link[]> => Map<string, Link[]>().withMutations(map => {
    links.forEach(link => {
      if (!link.plaidInstitutionId) {
        return;
      }

      if (map.has(link.plaidInstitutionId)) {
        map.get(link.plaidInstitutionId).push(link);
      } else {
        map.set(link.plaidInstitutionId, [
          link,
        ]);
      }
    });
  }));