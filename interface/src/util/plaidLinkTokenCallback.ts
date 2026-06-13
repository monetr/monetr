import type { WithJsonValues } from '@monetr/interface/util/json';
import request from '@monetr/interface/util/request';

export class LinkTokenCallbackResponse {
  linkId: number;

  constructor(data: WithJsonValues<LinkTokenCallbackResponse>) {
    this.linkId = data.linkId;
  }
}

export default function plaidLinkTokenCallback(
  publicToken: string,
  institutionId: string,
  institutionName: string,
  accountIds: string[],
): Promise<LinkTokenCallbackResponse> {
  return request<WithJsonValues<LinkTokenCallbackResponse>>({
    method: 'POST',
    url: '/api/plaid/link/token/callback',
    data: {
      publicToken,
      institutionId,
      institutionName,
      accountIds,
    },
  }).then(result => new LinkTokenCallbackResponse(result.data));
}
