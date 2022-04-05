# Authentication

Authentication resources for monetr's REST API.

## Login with an email and password

Provide login credentials to authenticate a user. This API will not respond with a token in the body. It stores the
token in an HTTP only cookie to prevent it from being accessible from Javascript code in the browser. If the credentials
are valid then you will receive a `200 OK` status code in the response.

```http title="HTTP"
POST /api/authentication/login
```

| Attribute  | Type   | Required | Description                                                                                                                                                   |
|------------|--------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `email`    | string | yes      | The email address associated with the desired login.                                                                                                          |
| `password` | string | yes      | The login password.                                                                                                                                           |
| `captcha`  | string | no *     | By default this field is not required, but if captcha is enabled then this field will be required. It should be the resulting value of a ReCAPTCHA challenge. |
| `totp`     | string | no       | Not yet fully implemented, but will be used to provide TOTP codes for the user during the authentication flow.                                                |

### Login Examples

```shell title="Example Login Request"
curl --request POST \
  --url "https://my.monetr.app/api/authentication/login" \
  --header "content-type: application/json" \
  --data '{
    "email": "email@example.com",
    "password": "superSecureP@ssw0rd"
}'
```

```json title="Example Response"
{
  "isActive": true
}
```

When the request is successful the response body is pretty minimal. User details should be retrieved using a follow up
request to [Get Me](user.md#get-me-current-user)

### Login Errors

If the credentials provided are not valid you will receive the following response body:

```json title="Invalid Credentials"
{
  "error": "invalid email and password"
}
```



