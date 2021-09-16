# Email Verification

This package handles a lot of the logic around email verification. Because we provide the user's email address to Plaid
and Stripe for connecting bank accounts and billing purposes we need to make sure that the user does own the provided
email.

## Workflow

1. The user supplies their email during the Sign Up form in the UI.
    a. We verify that another login with the same email does not already exist.
    b. We create a stripe customer with the email provided.
2. We then call our `CreateEmailVerificationToken` providing it the address.
    a. This method will generate a signed token (JWT or anything really) that we can verify. This token should have a
       short lived expiration. Something like 10 minutes.
    b. The user receives an email with a link that includes the token. Upon following the link we validate that token.
3. `ValidateEmailVerificationToken` is called, it receives the token provided by the link. The token has the user's
   email address as well as an expiration encoded in it. We look up the login with that email address.
    a. If that login is already verified, we return an error indicating that the link is not valid.
    b. If that login is not verified we set it to verified, but only if the token is not expired.
4. We present the user with a toast saying their email has been verified, and prompt them to re-enter credentials to
   login.

## Tokens

Tokens right now are meant to change over time. They are generated using the `EmailVerificationTokenGenerator` interface
and can essentially be implemented however it is needed. Initially they'll be implemented using JWT but in the future I
want to add support for other types of tokens as well that might be more secure.