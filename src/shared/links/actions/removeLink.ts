import { Dispatch } from "redux";
import { RemoveLink } from "shared/links/actions";
import request from "shared/util/request";
import Link from "data/Link";
import { getBankAccountsByLinkId } from "shared/bankAccounts/selectors/getBankAccountsByLinkId";

interface ActionWithState {
  (dispatch: Dispatch, getState: () => any): Promise<void>
}

export const removeLinkRequest = {
  type: RemoveLink.Request,
};

export const removeLinkFailure = {
  type: RemoveLink.Failure,
};

export default function removeLink(link: Link): ActionWithState {
  return (dispatch: Dispatch, getState) => {
    dispatch(removeLinkRequest);

    const bankAccounts = getBankAccountsByLinkId(link.linkId)(getState());

    return request()
      .delete(`/links/${ link.linkId }`)
      .then(() => {
        dispatch({
          type: RemoveLink.Success,
          payload: {
            link,
            bankAccounts: bankAccounts.valueSeq().toArray(),
          }
        });
      })
      .catch(error => {
        dispatch(removeLinkFailure);

        throw error;
      });
  };
}
