import { FETCH_LINKS_FAILURE, FETCH_LINKS_REQUEST, FETCH_LINKS_SUCCESS } from "shared/links/actions";
import request from "shared/util/request";

import { Map } from 'immutable';
import Link from "data/Link";


export const fetchLinksRequest = {
  type: FETCH_LINKS_REQUEST,
};

export const fetchLinksFailure = {
  type: FETCH_LINKS_FAILURE,
};

export default function fetchLinks() {
  return dispatch => {
    dispatch(fetchLinksRequest);
    return request().get('/api/links')
      .then(result => {
        dispatch({
          type: FETCH_LINKS_SUCCESS,
          payload: Map<number, Link>().withMutations(map => {
            (result.data.links || []).forEach((link: Link) =>
              map.set(link.linkId, new Link(link))
            );
          }),
        });
      })
      .catch(error => {
        dispatch(fetchLinksFailure);
        throw error;
      });
  }
}
