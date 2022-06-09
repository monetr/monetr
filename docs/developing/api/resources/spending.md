# Spending

Spending objects are used to represent how the user wants to spend their money and how frequently they want to or need
to spend it. Spending objects create an ["earmark"](https://www.merriam-webster.com/dictionary/earmark) against the
available balance of the bank account they are associated with. This is then used to allow monetr to calculate an amount
that is safe for the user to spend at any given time; while making sure they still have funds for their defined
financial obligations.

## List Spending

This endpoint does not support any pagination, it simply returns all of the spending objects for the provided bank
account ID.

```http title="HTTP"
GET /api/bank_accounts/{bankAccountId}/spending
```

### Request Path

| Attribute         | Type     | Required   | Description                                                              |
| ----------------- | -------- | ---------- | ------------------------------------------------------------------------ |
| `bankAccountId`   | number   | yes        | The ID of the bank account the spending objects belong to.               |

### Response Body

| Attribute                | Type     | Required | Description                                                                                                                                                                                                                                                                              |
| ----                     | ----     | ----     | ----                                                                                                                                                                                                                                                                                     |
| `spendingId`             | number   | yes      | The unique identifier for a spending object within monetr.                                                                                                                                                                                                                               |
| `bankAccountId`          | number   | yes      | The bank account that the spending object belongs to. This will match the path parameter for this API endpoint.                                                                                                                                                                          |
| `fundingScheduleId`      | number   | yes      | The ID of the funding schedule that is used to calculate contributions to the spending object.                                                                                                                                                                                           |
| `name`                   | string   | yes      | The name or title of the spending object, this must be unique within a bank account and `spendingType`.                                                                                                                                                                                  |
| `description`            | string   | no       | The description for the spending object.                                                                                                                                                                                                                                                 |
| `spendingType`           | enum     | yes      | The type of spending object this is. <br> - `0` Expense <br> - `1` Goal                                                                                                                                                                                                                  |
| `targetAmount`           | number   | yes      | The amount of money (in cents) that this spending object needs at or before it's `nextRecurrence` date.                                                                                                                                                                                  |
| `currentAmount`          | number   | yes      | The amount of money (in cents) that is allocated to this spending object.                                                                                                                                                                                                                |
| `usedAmount`             | number   | no       | This field is only used for Goals, it is used to keep track of how much has been spent from the goal. When a transaction is spent from a Goal it increments the `usedAmount` equal the the amount of the transaction or the `currentAmount` of the spending object, whichever is lesser. |
| `recurrenceRule`         | string   | no       | The RRule used to calculate the due dates for a spending object, this is only present for Expenses.                                                                                                                                                                                      |
| `lastRecurrence`         | datetime | no       | The timestamp of the last time this spending object was due. This is not a reflection of the last time the spending object was used.                                                                                                                                                     |
| `nextRecurrence`         | datetime | yes      | The timestamp of the next time this spending object is due, for a Goal this is just the day you want to complete the Goal. For an expense this is the next time it will recur before it is recalculated.                                                                                 |
| `nextContributionAmount` | number   | yes      | The amount of money (in cents) that will be allocated to the `currentAmount` the next time the funding schedule is processed.                                                                                                                                                            |
| `isBehind`               | boolean  | yes      | Spending objects can fall behind if there will not be a funding event before the `nextRecurrence` and the `currentAmount` is less than the `targetAmount`.                                                                                                                               |
| `isPaused`               | boolean  | no       | Spending objects can be paused, when they are paused they will not be funded or updated automatically.                                                                                                                                                                                   |
| `dateCreated`            | datetime | yes      | The timestamp of when this spending object was created.                                                                                                                                                                                                                                  |

