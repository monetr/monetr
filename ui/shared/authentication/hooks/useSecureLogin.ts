import srp from '@elliotcourant/srp.js';
import { useNavigate } from 'react-router-dom';
import useBootstrapLogin from 'shared/authentication/actions/bootstrapLogin';
import request from 'shared/util/request';
import { bigIntToHex, hexToBigInt, hexToUint8Array, uint8ArrayToHex } from 'util/math';

export interface LoginParameters {
  email: string;
  password: string;
  captcha?: string | null;
}

export default function useSecureLogin(): (params: LoginParameters) => Promise<void> {
  const navigate = useNavigate();
  const bootstrapLogin = useBootstrapLogin();

  interface Challenge {
    salt?: Uint8Array | null;
    public?: bigint | null;
    secure: boolean;
  }

  return async (params: LoginParameters): Promise<void> => {
    const authenticationSessionId = uint8ArrayToHex(crypto.getRandomValues(new Uint8Array(32)));

    async function authenticateLegacy(): Promise<void> {
      return request().post('/authentication/login', params)
        .then(result => {
          // Then bootstrap the authentication, once it's bootstrapped we want to consider the `nextUrl` field from the
          // login response above. If the nextUrl is present, then we want to navigate the user to that path. If it is not
          // present then we can direct the user to the root path.
          return bootstrapLogin().then(() => navigate(result?.data?.nextUrl || '/'));
        })
        .catch(error => {
          // More important than the message though is the status of the response. If the status code was 428 then that
          // means the credentials are valid, but the user has not verified their email yet. If this is the case we want
          // to redirect them to the resend email verification page and autofill that user's email address.
          switch (error?.response?.status) {
            case 428: // Email not verified.
              switch (error?.response?.data?.code) {
                case 'MFA_REQUIRED':
                  return navigate('/login/mfa', {
                    state: {
                      'emailAddress': params.email,
                      'password': params.password,
                      // TODO ReCAPTCHA?
                    }
                  });
                case 'EMAIL_NOT_VERIFIED':
                  return navigate('/verify/email/resend', {
                    state: {
                      'emailAddress': params.email,
                    }
                  });
                default:
                  throw error;
              }
            case 403: // Invalid login.
              throw error;
          }

          throw error;
        });
    }

    async function authenticateSecure(challenge: Challenge): Promise<void> {
      const email = params.email.trim().toLowerCase();
      const x = await srp.KDFSHA512(challenge.salt!, email, params.password);

      const client = new srp.SRP(srp.G8192);
      await client.Setup(srp.Mode.Client, x, null);
      // This will set B since we are the client.
      client.SetOthersPublic(challenge.public!);
      // We need to establish our public before generating M.
      const A = await client.EphemeralPublic();
      await client.Key();
      const clientProof = await client.M(challenge.salt!, email);

      return request().post(`/authentication/secure/authenticate`, {
        public: bigIntToHex(A),
        proof: uint8ArrayToHex(clientProof),
        sessionId: authenticationSessionId,
      })
        .then(exchangeResponse => client.GoodServerProof(
          challenge.salt!,
          email,
          hexToUint8Array(exchangeResponse.data.proof),
        ))
        .then(good => {
          if (!good) throw new Error('invalid credentials');
        })
        .then(() => bootstrapLogin())
        .then(() => navigate('/'));
      // TODO implement nextUrl with SRP.
      // TODO implement MFA
      // TODO implement email verification.
      // Caller needs to catch.
    }

    const email = params.email.trim().toLowerCase();

    return request().post(`/authentication/secure/challenge`, {
      email: email,
      sessionId: authenticationSessionId,
    })
      .then(challengeResponse => ({
        salt: challengeResponse.data.salt && hexToUint8Array(challengeResponse.data.salt),
        public: challengeResponse.data.public && hexToBigInt(challengeResponse.data.public),
        secure: challengeResponse.data.secure,
      }))
      // If secure remote password is available for the user, then continue to authenticate using that. Otherwise, fall
      // back to legacy authentication.
      .then(challenge => challenge.secure ? authenticateSecure(challenge) : authenticateLegacy())
      .catch(error => {
        console.warn(error);
        throw new Error('invalid credentials provided');
      });
  }
}
