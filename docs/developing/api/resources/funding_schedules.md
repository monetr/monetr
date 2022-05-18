# Funding Schedules

Funding schedules are used to tell monetr when to allocate funds to spending objects. They represent the frequency that
you are paid; or the frequency that you would like to allocate funds to things you are budgeting for. Funding schedules
are specific to a single bank account and must have a unique name within a bank account.

## List Funding Schedules

This endpoint does not support any pagination, it simply returns all of the funding schedules for the provided bank
account Id.

```http title="HTTP"
GET /api/bank_accounts/{bankAccountId}/funding_schedules
```

### Request Path

| Attribute       | Type   | Required | Description                                                            |
|-----------------|--------|----------|------------------------------------------------------------------------|
| `bankAccountId` | number | yes      | The ID of the bank account this new funding schedule should belong to. |

### Response Body

| Attribute           | Type     | Required | Description                                                                                                                                                                  |
|---------------------|----------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `fundingScheduleId` | number   | yes      | The globally unique identifier for a funding schedule.                                                                                                                       |
| `bankAccountId`     | number   | yes      | The bank account that this funding schedule belongs to, this will be the same as the `bankAccountId` provided in the URL path.                                               |
| `name`              | string   | yes      | The name of the funding schedule; this is unique per bank account.                                                                                                           |
| `description`       | string   | no       | If the funding schedule was created by the UI, then this will be a summary of the `rule` field in english.                                                                   |
| `rule`              | string   | yes      | The RRule provided when the funding schedule was created; representing how frequently this funding schedule will occur.                                                      |
| `excludeWeekends`   | bool     | yes      | If this is true, then the `nextOccurrence` will never fall on a weekend. If the `rule` would have it fall on a weekend, then the date is adjusted to be the previous Friday. |
| `nextOccurrence`    | datetime | yes      | The next time this funding schedule will allocate funds to it's spending objects.                                                                                            |

### List Funding Schedules Examples

```shell title="Example List Funding Schedules Request"
curl --request GET \
  --url "https://my.monetr.app/api/bank_accounts/123/funding_schedules"
```

#### Successful

```json title="200 Ok"
[
    {
      "fundingScheduleId": 44,
      "bankAccountId": 123,
      "name": "Payday",
      "description": "The 15th and the Last day of every month",
      "rule": "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
      "excludeWeekends": false,
      "nextOccurrence": "2022-05-31T00:00:00-06:00"
    }
]
```

## Create Funding Schedule

Create a funding schedule by providing some basic information about when the funding schedule will occur next as well as
how frequently it occurs.

```http title="HTTP"
POST /api/bank_accounts/{bankAccountId}/funding_schedules
```

### Request Path

| Attribute       | Type   | Required | Description                                                            |
|-----------------|--------|----------|------------------------------------------------------------------------|
| `bankAccountId` | number | yes      | The ID of the bank account this new funding schedule should belong to. |

### Request Body

| Attribute         | Type     | Required | Description                                                                                                                                                                                                                                                                                                                                                                                                          |
|-------------------|----------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name`            | string   | yes      | The name of the new funding schedule. For a given `bankAccountId` this value must be unique.                                                                                                                                                                                                                                                                                                                         |
| `rule`            | string   | yes      | The RRule representing how the funding schedule should occur. See [RFC5545](https://datatracker.ietf.org/doc/html/rfc5545).<br> **Examples**: <br> - The 15th and last day of every month: `FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1` <br> - Every other friday: `FREQ=WEEKLY;INTERVAL=2;BYDAY=FR` <br> - Every friday: `FREQ=WEEKLY;INTERVAL=1;BYDAY=FR`                                                            |
| `description`     | string   | no       | The description of the funding schedule, this can be anything you want; but in the UI this is auto filled to be a string of the RRule. For example; `The 15th and Last day of every month`                                                                                                                                                                                                                           |
| `excludeWeekends` | bool     | no       | :fontawesome-solid-flask: Exclude weekends will adjust the occurrence dates if they were to fall on a weekend. It will set the date to be the closest previous business day. This will not impact the frequency of the funding schedule.                                                                                                                                                                             |
| `nextOccurrence`  | datetime | no       | You can provide the next occurrence date in the create request, this date will be used on subsequence recurrences to determine the following contribution dates. For rules that have static dates defined like the 15th and last day of the month, this will not affect subsequent recurrences. But for rules that can be more loose, like every other friday; this will determine which "every other" friday it is. |

### Response Body

| Attribute           | Type     | Required | Description                                                                                                                                                                                                         |
|---------------------|----------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `fundingScheduleId` | number   | yes      | The unique identifier for the funding schedule you created. This is globally unique within monetr.                                                                                                                  |
| `bankAccountId`     | number   | yes      | This will be the value of the `bankAccountId` parameter you provided in the API path, and is included on `GET` requests as well.                                                                                    |
| `name`              | string   | yes      | The name of the funding schedule provided by you, if there were leading or trailing spaces they will have been trimmed.                                                                                             |
| `description`       | string   | no       | If a description was provided then it will be present here.                                                                                                                                                         |
| `rule`              | string   | yes      | The RRule provided when the funding schedule was created.                                                                                                                                                           |
| `excludeWeekends`   | bool     | yes      | If a value was provided in the create request, it will be present here; otherwise this will be `false`.                                                                                                             |
| `nextOccurrence`    | datetime | yes      | If you provided a date time and it was in the future, that will be <br/>used for the next occurrence. If one was not provided, then this value will be calculated using the `rule` field and the current timestamp. |

### Create Funding Schedule Examples

```shell title="Example Create Funding Schedule Request"
curl --request POST \
  --url "https://my.monetr.app/api/bank_accounts/123/funding_schedules" \
  --header "content-type: application/json" \
  --data '{
    "name": "Payday",
    "description": "The 15th and Last day of every month",
    "rule": "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
    "excludeWeekends": false,
    "nextOccurrence": "2022-05-31T00:00:00-06:00"
}'
```

#### Successful

If the funding schedule was created successfully, then you'll receive the created object back.

```json title="200 Ok"
{
  "fundingScheduleId": 44,
  "bankAccountId": 123,
  "name": "Payday",
  "description": "The 15th and the Last day of every month",
  "rule": "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
  "excludeWeekends": false,
  "nextOccurrence": "2022-05-31T00:00:00-06:00"
}
```

## Delete Funding Schedule

Delete a funding schedule for a given bank account. This can only be done if there are no spending objects associated
with the funding schedule you want to remove. This will simply return a `200 Ok` status code if it succeeds, there is
no response body.

```http title="HTTP"
DELETE /api/bank_accounts/{bankAccountId}/funding_schedules/{fundingScheduleId}
```

### Request Path

| Attribute           | Type     | Required   | Description                                                              |
| -----------------   | -------- | ---------- | ------------------------------------------------------------------------ |
| `bankAccountId`     | number   | yes        | The ID of the bank account the funding schedule belongs to.              |
| `fundingScheduleId` | number   | yes        | The ID of the funding schedule you want to remove.                       |


### Delete Funding Schedule Example

```shell title="Example Delete Funding Schedule Request"
curl --request DELETE \
  --url "https://my.monetr.app/api/bank_accounts/123/funding_schedules/123"
```

#### Funding Schedule Does Not Exist

If you try to delete a funding schedule that does not exist, then you will receive the following error.

```json title="404 Not Found"
{
  "error": "cannot remove funding schedule, it does not exist"
}
```
