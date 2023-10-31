# Links

Links are used to represent a connection between monetr and Plaid (or potentially some other source of data for bank
transactions and balances). But links can also be used to represent a manual budgeting "area". Links are either created
automatically by following the Plaid workflow; or can be created manually using these API endpoints to do manual
budgeting.

## List Links

This endpoint does not support any pagination, it simply returns all of the links associated with the currently
authenticated account.

```http title="HTTP"
GET /api/links
```

### Response Body

| Attribute               | Type     | Required | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ----                    | ----     | ----     | ----                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| `linkId`                | number   | yes      | The unique identifier for a given link within monetr. This is unique.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `linkType`              | enum     | yes      | What type of link this is. <br>- `0` Unknown (likely an error state) <br>- `1` Plaid Link (**managed**) <br>- `2` Manual Link                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| `plaidInstitutionId`    | string   | no       | If the link is a Plaid link, then this will include Plaid's unique identifier for the institution it is assocaited with. This is used to retrieve basic institution information. It is `null` for manual links                                                                                                                                                                                                                                                                                                                                                     |
| `linkStatus`            | enum     | yes      | The status of the link, for manual links this will always be `2`. <br>- `0` Unknown (likely an error state) <br>- `1` Pending (the link is currently being setup by the user, or waiting on Plaid data) <br>- `2` Setup (the link is healthy and functioning normally) <br>- `3` Error (something is wrong with the link that might require user interaction to resolve) <br>- `4` Pending Expiration (the credentials for the link are going to expire and it needs to be reauthenticated) <br> - `5` Revoked (the user has revoked access to the bank via Plaid) |
| `errorCode`             | string   | no       | If the link is an error state this will have an error code from Plaid, this is used to help decide what action needs to be taken to fix the link.                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `expirationDate`        | datetime | no       | If the Plaid link can expire, then its expiration date will be stored here. After this date the link will need to be reauthenticated.                                                                                                                                                                                                                                                                                                                                                                                                                              |
| `institutionName`       | string   | yes      | The name of the instituion or link. If this is a Plaid link then this will be the name of the bank assocaited with the link in Plaid. If this is a manual link then this will be the name given to the link at creation.                                                                                                                                                                                                                                                                                                                                           |
| `customInstitutionName` | string   | no       | Plaid links cannot have their `institutionName` updated, but if the user wants a custom name for a Plaid link then this field will be used.                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `createdAt`             | datetime | yes      | When this link was created. This will be in UTC.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `createdByUserId`       | number   | yes      | The ID of the user who created the link. Links are "owned" by the user who creates them. Even if multiple people have access to the same account, a link is "owned" by the one who creates it.                                                                                                                                                                                                                                                                                                                                                                     |
| `updatedAt`             | datetime | yes      | When this link was last updated. **Note**: This field might not be well maintained and should not be relied on, it may be deprecated in the future.                                                                                                                                                                                                                                                                                                                                                                                                                |
| `lastManualSync`        | datetime | no       | The last time this link was "manually" triggered to sync.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `lastSuccessfulUpdate`  | datetime | no       | The last time this link successfully received new data from Plaid.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |

## Create (Manual) Link

Manual links can be created witha a post request. But allow only a subset of the fields to be provided and managed by
the client.

### Request Body

| Attribute               | Type   | Required | Description                                                                                           |
| ----                    | ----   | ----     | ----                                                                                                  |
| `institutionName`       | string | yes      | The name of the instituion or link.                                                                   |
| `customInstitutionName` | string | no       | Custom institution name for a link, this field may be subject to changes in behavior for manual links |

### Response Body

| Attribute               | Type     | Required | Description                                                                                                                                                                                                              |
| ----                    | ----     | ----     | ----                                                                                                                                                                                                                     |
| `linkId`                | number   | yes      | The unique identifier for a given link within monetr. This is unique.                                                                                                                                                    |
| `linkType`              | enum     | yes      | Will always be `2` for manual links.                                                                                                                                                                                     |
| `linkStatus`            | enum     | yes      | The status of the link, for manual links this will always be `2`.                                                                                                                                                        |
| `institutionName`       | string   | yes      | The name of the instituion or link. If this is a Plaid link then this will be the name of the bank assocaited with the link in Plaid. If this is a manual link then this will be the name given to the link at creation. |
| `customInstitutionName` | string   | no       | Plaid links cannot have their `institutionName` updated, but if the user wants a custom name for a Plaid link then this field will be used.                                                                              |
| `createdAt`             | datetime | yes      | When this link was created. This will be in UTC.                                                                                                                                                                         |
| `createdByUserId`       | number   | yes      | The ID of the user who created the link. Links are "owned" by the user who creates them. Even if multiple people have access to the same account, a link is "owned" by the one who creates it.                           |
| `updatedAt`             | datetime | yes      | When this link was last updated. **Note**: This field might not be well maintained and should not be relied on, it may be deprecated in the future.                                                                      |

### Create Manual Link Example

```shell title="Example Create Manual Link Request"
curl --request POST \
  --url "https://my.monetr.app/api/links" \
  --header "content-type: application/json" \
  --data '{
    "institutionName": "My Manual Budgeting"
}'
```


