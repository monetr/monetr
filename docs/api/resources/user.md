# User

User resources for monetr's REST API.

## Get Me (Current User)

Retrieve information about the currently authenticated user. In the UI this API endpoint is also used to determine if
there is a currently authenticated user. Because authentication is stored in HTTP only cookies there is no way for the
UI to see if the cookie is present, so upon bootstrapping it makes an API call to this endpoint to determine if there is
an authenticated user.

```http title="HTTP"
GET /api/users/me
```
