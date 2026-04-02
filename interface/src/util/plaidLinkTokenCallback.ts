import request from '@monetr/interface/util/request';

export class LinkTokenCallbackResponse {
  linkId: number;

  constructor(data: Partial<LinkTokenCallbackResponse>) {
    Object.assign(this, data);
  }
}

export default function plaidLinkTokenCallback(
  publicToken: string,
  institutionId: string,
  institutionName: string,
  accountIds: string[],
): Promise<LinkTokenCallbackResponse> {
  return request<Partial<LinkTokenCallbackResponse>>({ method: 'POST', url: '/api/plaid/link/token/callback', data: {
      publicToken,
      institutionId,
      institutionName,
      accountIds,
    } })
    .then(
      result =>
        new LinkTokenCallbackResponse({
          linkId: result.data.linkId,
        }),
    );
}
