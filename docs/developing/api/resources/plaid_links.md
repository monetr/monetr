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

[^1]:

    monetr can be configured to limit the number of links an account can have. This is for those using development
    credentials. If you intend to self-host monetr and let friends and/or family use it, it is recommended that you
    limit how many links someone can have to prevent someone from adding many links to their account and using up your
    limited number of development links.

[^2]:

    If you fully authenticate a bank account, but you never exchange the `public_token` for that authenticated bank
    account. It _should not_ affect the number of accounts linked for development credentials.
