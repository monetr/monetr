import srp from '@elliotcourant/srp.js';
import User from 'models/User';
import request from 'shared/util/request';
import { bigIntToHex, uint8ArrayToHex } from 'util/math';

export interface RegisterParameters {
  agree: boolean;
  betaCode?: string | null;
  captcha?: string | null;
  email: string;
  firstName: string;
  lastName: string;
  password: string;
  timezone: string;
}

export interface RegisterReponse {
  isActive: boolean;
  message: string | null;
  nextUrl: string | null;
  requireVerification: boolean | null;
  user: User | null;
}

export default function useRegister(): (params: RegisterParameters) => Promise<RegisterReponse> {
  return async (params: RegisterParameters): Promise<RegisterReponse> => {
    let salt = new Uint8Array(32);
    salt = crypto.getRandomValues(salt);

    const email = params.email.trim().toLowerCase();

    const x = await srp.KDFSHA512(salt, email, params.password);

    const client = new srp.SRP(srp.G8192);
    await client.Setup(srp.Mode.Client, x, null);

    return request().post(`/authentication/secure/register`, {
      'email': email,
      'firstName': params.firstName,
      'lastName': params.lastName,
      'timezone': params.timezone,
      'agree': params.agree,
      'verifier': bigIntToHex(client.Verifier()),
      'salt': uint8ArrayToHex(salt),
      // If ReCAPTCHA was provided, then pass that along as well.
      ...(params.captcha && { captcha: params.captcha }),
      // If a beta code was provided, then include that too.
      ...(params.betaCode && { betaCode: params.betaCode })
    })
      .then(result => result.data as RegisterReponse)
  }
}
