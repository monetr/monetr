import { Dispatch } from "redux";
import Link from "data/Link";
import request from "shared/util/request";
import { CreateLinks } from "shared/links/actions";


interface ActionWithState {
  (dispatch: Dispatch): Promise<Link>
}

export default function createLink(link: Link): ActionWithState {
  return (dispatch: Dispatch) => {
    dispatch({
      type: CreateLinks.Request
    });

    return request()
      .post(`/links`, link)
      .then(result => {
        const link = new Link(result.data);
        dispatch({
          type: CreateLinks.Success,
          payload: link,
        });

        return link;
      })
      .catch(error => {
        dispatch({
          type: CreateLinks.Failure
        });

        throw error;
      });
  }
}
