# Account Settings

At the moment settings are an account level object.

## Get Settings

You can retrieve settings for the currently authenticated account using this endpoint.

```http title="HTTP"
GET /api/account/settings
```

### Response Body

| Attribute        | Type   | Required | Description                                                                                                                                                                   |
|------------------|--------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `maxSafeToSpend` | object | yes      | A sub-object outlining the settings for configuring a maxium allowed safe to spend per period between funding schedules. **This feature has not been completely implemented** |
