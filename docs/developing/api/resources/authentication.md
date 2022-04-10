# Authentication

Authentication resources for monetr's REST API.

Authentication is primarily broken up into three objects. Logins, users, and accounts. Each of these objects are
different from each other, but all of them are related. Logins are used to represent a single person's credentials for
accessing monetr, as well as some of their basic information like their name. Users are a child object of logins, and a
single login can have multiple users. Users are tied to a single account. Accounts are how all data is separated inside
the application. When you create a Plaid Link or expenses, or other budgeting items. They are created at an account
level. Your login tells us what users represent you, and those users tell us what accounts you have access to.

``` mermaid
classDiagram
    Login --> "many" User : Can have
    User "1" --> "1" Account : Associated with
    class Login{
      +int loginId
      +String email
      +String firstName
      +String lastName
    }
    class User{
      +int userId
      +int loginId
      +int accountId
    }
    class Account{
      +int accountId
    }
```

??? note

    At the moment logins are limited to a single user. This is a software constraint for now. The design to allow
    logins access to multiple users was to allow (in the future) people to have shared access to an account. Such as
    with a spouse.

## Login

Provide login credentials to authenticate a user. This API will not respond with a token in the body. It stores the
token in an HTTP only cookie to prevent it from being accessible from Javascript code in the browser. If the credentials
are valid then you will receive a `200 OK` status code in the response.

```http title="HTTP"
POST /api/authentication/login
```

### Request Body

| Attribute  | Type   | Required | Description                                                                                                                                                   |
|------------|--------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `email`    | string | yes      | The email address associated with the desired login.                                                                                                          |
| `password` | string | yes      | The login password. Passwords must be at least 8 characters.                                                                                                  |
| `captcha`  | string | no *     | By default this field is not required, but if captcha is enabled then this field will be required. It should be the resulting value of a ReCAPTCHA challenge. |
| `totp`     | string | no       | Not yet fully implemented, but will be used to provide TOTP codes for the user during the authentication flow.                                                |

### Response Body

| Attribute  | Type   | Required | Description                                                                                                                                                             |
|------------|--------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `isActive` | bool   | yes      | Used as an indicator for whether or not the current user's subscription is active. If billing is disabled this is always true.                                          |
| `nextUrl`  | string | no       | If the API needs to direct a user to a certain path after authenticating, this field will be present. This is intended to be used for on-boarding or for billing flows. |

??? note

    When the request is successful the response body is pretty minimal. User details should be retrieved using a follow 
    up request to [Get Me](user.md#get-me-current-user)

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

#### Successful

If the credentials provided are valid and there is nothing else to be done, then the response will simply be this.

```json title="200 Ok"
{
  "isActive": true
}
```

#### Subscription is not active

If the subscription for the authenticated user is not active (or if there is no subscription at all), then the response
body will contain a path indicating to the UI what the next URL should be.

```json title="200 Ok"
{
  "isActive": false,
  "nextUrl": "/account/subscribe"
}
```

### Login Errors

#### Invalid Credentials

If the credentials provided are not valid you will receive the following response body:

```json title="401 Unauthorized"
{
  "error": "invalid email and password"
}
```

#### Email is not verified

If email verification is required by the server, then it is possible to get a login failure response even with valid
credentials. If the credentials provided are valid, but the login's email is not verified yet; then you will receive the
following response.

```json title="428 Precondition Required"
{
  "error": "email address is not verified",
  "code": "EMAIL_NOT_VERIFIED"
}
```

#### MFA is required

That is not the only error that can result in a `428` status code, if the user requires MFA you will receive the same
status code upon providing valid credentials. But the code in the error body will be different and represent what action
needs to be taken by the user.

```json title="428 Precondition Required"
{
  "error": "login requires MFA",
  "code": "MFA_REQUIRED"
}
```

## Logout

Because cookies are HTTP only, there is no way to remove the cookies from our UI code. Instead, we have a logout
endpoint that removes the cookies.

```http title="HTTP"
GET /api/authentication/logout
```

### Logout Examples

```shell title="Example Login Request"
curl --request GET \
  --url "https://my.monetr.app/api/authentication/logout"
```

Logout does not return an error if the cookie is not present, it will always return a `200` status code with an empty
response body.

## Register

New users can register for monetr using this endpoint. It can be configured to require ReCAPTCHA to reduce the
likelihood that the endpoint will be spammed. Even in self-hosted deployments, it will require a valid email address is
used. [^1] Registering will create a new login, user, and account using the provided details.

If billing is enabled on the server, your email address will be used to create a Stripe customer. Stripe is used to
manage subscriptions and this way Stripe has a way to contact you or vice versa if needed.

```http title="HTTP"
POST /api/authentication/register
```

### Request Body

| Attribute   | Type   | Required | Description                                                                                                                                                   |
|-------------|--------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `email`     | string | yes      | The email address associated with the desired login.                                                                                                          |
| `password`  | string | yes      | The login password. Passwords must be at least 8 characters. Leading and trailing spaces are trimmed from the password.                                       |
| `firstName` | string | yes      | By default this field is not required, but if captcha is enabled then this field will be required. It should be the resulting value of a ReCAPTCHA challenge. |
| `lastName`  | string | yes      | Not yet fully implemented, but will be used to provide TOTP codes for the user during the authentication flow.                                                |
| `timezone`  | string | yes      | The timezone you want your account to be configured for.                                                                                                      |
| `captcha`   | string | no *     | Can be required if ReCAPTCHA is enabled on the server.                                                                                                        |
| `betaCode`  | string | no *     | Can be required if the server requires an access code to create an account.                                                                                   |
| `agree`     | bool   | yes      | Used to denote that the user has agreed to the terms of use for monetr.                                                                                       |

### Response Body

| Attribute             | Type   | Required | Description                                                                                                                                                                                                      |
|-----------------------|--------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `message`             | string | no       | If the user needs to verify their email then this will contain a message that can be presented to the user.                                                                                                      |
| `requireVerification` | bool   | yes      | If the server requires that the login's email is verified, the fields below this will be omitted in the response body and this will be true. If this is false then the user does not need to verify their email. |
| `nextUrl`             | string | no       | If the API requires that the user take some on-boarding action, such as setting up a subscription. There will be a path provided here for the user to be redirected to.                                          |
| `user`                | User   | no       | If email verification is not required by the server then the newly created user object will be present in the response.                                                                                          |
| `isActive`            | bool   | no       | Indicated whether or not the user is active, for servers that require billing this will be false initially.                                                                                                      |

### Register Examples

```shell title="Example Register Request"
curl --request POST \
  --url "https://my.monetr.app/api/authentication/register" \
  --header "content-type: application/json" \
  --data '{
    "email": "email@example.com",
    "password": "superSecureP@ssw0rd",
    "firstName": "Elliot",
    "lastName": "Courant",
    "timezone": "America/Chicago",
    "agree": true
}'
```

#### Successful

If the registration was successful and email verification is not necessary then you'll see a response body similar to
the following.

```json title="200 Ok"
{
  "nextUrl": "/setup",
  "user": {
    "loginId": 1234,
    "accountId": 1235,
    "userId": 1236,
    "login": {
      "loginId": 1234,
      "firstName": "Elliot",
      "lastName": "Courant",
    },
    "account": {
      "accountId": 1235,
      "timezone": "America/Chicago"
    }
  },
  "isActive": true,
  "requireVerification": false
}
```

#### Email verification required

If the registration succeeds, but we need to verify the login's email address.

```json title="200 Ok"
{
  "message": "A verification email has been sent to your email address, please verify your email.",
  "requireVerification": true
}
```

[^1]:

    Email addresses are not used to send people content unprompted. Email addresses provide a reliable way to assure
    uniqueness in users, as well as a way to contact them for things like billing and resetting forgotten passwords.
    For self-hosted deployments, it is not required that the email address used be one that can actually receive emails.
    However, this will limit your ability to easily reset forgotten passwords at this time.
