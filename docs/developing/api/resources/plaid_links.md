# Plaid Links

Plaid Links are supporting resources behind the normal [Link](links.md) resource. They provide a way to retrieve real
transaction and balance data from a person's bank account using Plaid's API. These endpoints expose some basic
functionality for interacting with Plaid Links through monetr's API rather than with Plaid directly.

## New Plaid Link Token

In order to connect a user's account with Plaid and thus their bank account, you need to create a link token. This token
is then used by the Plaid SDK in the UI in order to provide an authentication flow for that user and their bank.

Under the hood, monetr is making a call
to [Create Link Token](https://plaid.com/docs/api/tokens/#linktokencreate){:target="_blank"}.

```http title="HTTP"
GET /api/plaid/token/new
```

### Request Query Parameters

| Attribute   | Type | Required | Description                                                                                                                                                                                                                                                                                                                          |
|-------------|------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `use_cache` | bool | false    | If true then the endpoint will not try to create another link token if one has already been created for the current user and has not been completed. This is used to reduce the number of API calls made to Plaid. It is recommended to use `true` for this. The endpoint will never return a Link token that has already been used. |

### Response Body

| Attribute   | Type   | Required | Description                                                                                                   |
|-------------|--------|----------|---------------------------------------------------------------------------------------------------------------|
| `linkToken` | string | yes      | The link token created from Plaid, this can be used with the Plaid SDK to authenticate a user's bank account. |

### New Link Token Examples

```shell title="Example New Link Token Request"
curl --request GET \
  --url "https://my.monetr.app/api/plaid/token/new"
```

#### Successful

If the user has not reached their limit for links [^1] then they will receive a link token in the response and nothing
else.

```json title="200 Ok"
{
  "linkToken": "link-sandbox-af1a0311-da53-4636-b754-dd15cc058176"
}
```

### New Link Errors

#### Link Limit Reached

```json title="400 Bad Request"
{
  "error": "max number of Plaid links already reached"
}
```

#### Plaid API failure or other failure

```json title="500 Internal Server Error"
{
  "error": "failed to create link token"
}
```

## Plaid Token Callback

Once the user has completed authenticating their bank account using the Plaid SDK, you will be given a `public_token`.
This new token must be exchanged with Plaid's API in order to allow monetr to access the bank data. The bank account
should not be considered linked until this token is successfully exchanged[^2].

??? note

    :fontawesome-solid-server: Self-Hosted

    If webhooks are not enabled on the monetr server, then a job to retrieve transactions from Plaid will immediately be
    kicked off once the token has successfully been exchanged. This job will only attempt to retrieve the last 7 days of
    transactions from Plaid.

```http title="HTTP"
POST /api/plaid/token/callback
```

### Request Body

| Attribute         | Type     | Required | Description                                                                                                                                                |
|-------------------|----------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `publicToken`     | string   | yes      | The `public_token` you received from the Plaid Link flow via the SDK.                                                                                      |
| `institutionId`   | string   | yes      | The Plaid `institution_id`, the API will not return an error if this is not provided (TODO). But it is required in order for the API to function properly. |
| `institutionName` | string   | yes      | The name of the institution as provided by the Plaid SDK. This is used as the initial name for the link until the user decides to rename it.               |
| `accountIds`      | []string | yes      | An array of Plaid bank account IDs, this should be the list of IDs that the user selected when they were going through the link process.                   |

### Response Body

| Attribute | Type    | Required | Description                                                |
|-----------|---------|----------|------------------------------------------------------------|
| `linkId`  | integer | yes      | The ID of the newly created link for the Plaid connection. |

??? attention

    There is an additional field in the response body called `success`. This field is not present if the request fails
    and is only present if it succeeds. The status of the request is already indicated properly using the status code of
    the response. This field should not be used, it will be removed in the future.
    [![GitHub issue/pull request detail](https://img.shields.io/github/issues/detail/state/monetr/monetr/858?label=%23858%20-%20chore%3A%20Remove%20success%20field%20and%20any%20references%20from%20the%20%2Fapi%2Fplaid%2Ftoken%2Fcallback%20response.&logo=github)](https://github.com/monetr/monetr/issues/858){:target="_blank"}

### Plaid Token Callback Examples

```shell title="Example Plaid Token Callback Request"
curl --request POST \
  --url "https://my.monetr.app/api/plaid/token/callback" \
  --header "content-type: application/json" \
  --data '{
    "publicToken": "public-sandbox-5c224a01-8314-4491-a06f-39e193d5cddc",
    "institutionId": "ins_1",
    "institutionName": "A Banks Name",
    "accountIds": [
      "BxBXxLj1m4HMXBm9WZZmCWVbPjX16EHwv99vp",
      "lPNjeW1nR6CDn5okmGQ6hEpMo4lLNoSrzqDje"
    ]
}'
```

#### Successful

If monetr is able to exchange the public token successfully, then you will receive a successful response with the newly
created Link ID (1). <!--- This makes sure the annotation is not on this line, otherwise it does not work.-->
{ .annotate }

1. The Link ID is not a unique identifier for Plaid, it is monetr's internal unique identifier.

```json title="200 Ok"
{
  "linkId": 1234
}
```

### Plaid Token Callback Errors

#### No public token provided

You **must** provide a public token to create a Plaid link, there is no other way to authenticate a bank account through
Plaid.

```json title="400 Bad Request"
{
  "error": "must provide a public token"
}
```

#### No accounts provided

If you do not provide any Plaid bank account IDs then the API will not attempt to exchange the provided public token.

```json title="400 Bad Request"
{
  "error": "must select at least one account"
}
```

#### Plaid API failure or other failure

```json title="500 Internal Server Error"
{
  "error": "failed to exchange token"
}
```

## Update Plaid Link

Plaid links can be updated after they have been established. This can be used as a method of re-authenticating a link if
it ends up in an error state (though sometimes a link can end up in an error state without indicating [^3]). This can
also be used as a way to add additional accounts to a link if those accounts were not originally granted access when the
link was created.


```http title="HTTP"
PUT /api/plaid/link/update/{linkId}
```

### Request Path Parameters

| Attribute | Type   | Required | Description                                                                                                                                            |
| ----      | ----   | ----     | ----                                                                                                                                                   |
| `linkId`  | number | yes      | A link ID must be provided in order to put that link into update mode. The link must also be a Plaid link, a manual link will result in a bad request. |

### Request Query Parameters

| Attribute                  | Type    | Required | Description                                                                                                                                                                                                                                                 |
| ----                       | ----    | ----     | ----                                                                                                                                                                                                                                                        |
| `update_account_selection` | boolean | no       | This parameter is used to specify whether you want to put this link into an update mode that will allow you to add/remove accounts that are visible to monetr. This will change the Plaid Link dialog behavior slightly when it is presented to the client. |

### Update Link Examples

```shell title="Example Update Link Request"
curl --request PUT \
  --url "https://my.monetr.app/api/plaid/link/update/123?update_account_selection=true"
```

#### Successful

If the link is able to be put into update mode, then a Link Token is returned to the client. This can then be used with
the Plaid Link library to allow the user to update their account selection or re-authenticate their link with their
bank.

```json title="200 Ok"
{
  "linkToken": "link-sandbox-af1a0311-da53-4636-b754-dd15cc058176"
}
```

### Update Link Errors

#### Manual Link Update Requested

```json title="400 Bad Request"
{
  "error": "cannot update a non-Plaid link"
}
```

## Manually Resync Plaid Link

Sometimes you might need to manually trigger a resync with Plaid; this can happen if there were issues where a webhook
was not properly received. By triggering a manual resync, transactions for the last 14 days and balances for all
bank accounts within the specified link will be checked.

??? note

    This will not send a "sync" request to Plaid. This will only retrieve data already available via Plaid's API
    and update monetr's data accordingly.

??? attention

    This will not sync any removed transactions at this time. This will be resolved in the future by using the Plaid
    sync changes API, which will allow us to see all changes over a given period of time. But at the moment this will
    not update removed transactions.

    [![GitHub issue/pull request detail](https://img.shields.io/github/issues/detail/state/monetr/monetr/1270?label=bug%28plaid%29%3A%20Allow%20manually%20syncing%20to%20support%20removed%20transactions.&logo=github)](https://github.com/monetr/monetr/issues/1270){:target="_blank"}

```http title="HTTP"
POST /api/plaid/link/sync
```

### Request Body

| Attribute           | Type       | Required   | Description                                                                                                                                                          |
| ------------------- | ---------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------         |
| `linkId`            | number     | yes        | The link ID that you want to manually resync. This must be a Plaid link that is in a status of `Setup` or `Error`. Other link statuses will result in a bad request. |


### Manually Resync Examples

```shell title="Example manual resync request"
curl --request POST \
  --url "https://my.monetr.app/api/plaid/link/sync" \
  --header "content-type: application/json" \
  --data '{
    "linkId": 1234
}'
```

#### Successful

If the manual sync is kicked off successfully, you will recieve a `202 Accepted` status code with no response body.

### Errors

If the request fails then you will receive a JSON response body.

#### Provided Link does not exist

```json title="404 Not Found"
{
  "error": "failed to retrieve link: record does not exist"
}
```

#### Link is not a Plaid link

```json title="400 Bad Request"
{
  "error": "cannot manually sync a non-Plaid link"
}
```

#### Link is not in a valid status

```json title="400 Bad Request"
{
  "error": "link is not in a valid status, it cannot be manually synced"
}
```

#### Failed to enqueue sync job

```json title="500 Internal Server Error"
{
  "error": "failed to trigger manual sync"
}
```

[^1]:

    monetr can be configured to limit the number of links an account can have. This is for those using development
    credentials. If you intend to self-host monetr and let friends and/or family use it, it is recommended that you
    limit how many links someone can have to prevent someone from adding many links to their account and using up your
    limited number of development links.

[^2]:

    If you fully authenticate a bank account, but you never exchange the `public_token` for that authenticated bank
    account. It _should not_ affect the number of accounts linked for development credentials.


[^3]:

    While working on monetr I had a link fail to receive any updates from Plaid/the bank. It still showed that it was in
    a healthy state though. It didn't require re-authentication but at the time putting the link through "link update"
    was ultimately what resolved it.
