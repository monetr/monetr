import request from '@monetr/interface/util/request';

export interface SignUpArguments {
  betaCode: string | null;
  email: string;
  firstName: string;
  lastName: string;
  password: string;
  timezone: string;
  locale: string;
  challenge?: string;
  nonce?: number;
}

export interface SignUpResponse {
  message: string | null;
  nextUrl: string | null;
  requireVerification: boolean | null;
}

export default function useSignUp(): (args: SignUpArguments) => Promise<SignUpResponse> {
  return async (args: SignUpArguments) => {
    return await request({ method: 'POST', url: '/api/authentication/register', data: args }).then(
      result => result.data as SignUpResponse,
    );
  };
}
