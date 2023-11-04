import request from '@monetr/interface/util/request';

export interface SignUpArguments {
  betaCode: string | null;
  captcha: string | null;
  email: string;
  firstName: string;
  lastName: string;
  password: string;
  timezone: string;
}

export interface SignUpResponse {
  message: string | null;
  nextUrl: string | null;
  requireVerification: boolean | null;
}

export interface SignUpError {
  error: string;
}

export default function useSignUp(): (args: SignUpArguments) => Promise<SignUpResponse | SignUpError> {
  return async (args: SignUpArguments) => {
    return request().post('/authentication/register', args)
      .then(result => result.data as SignUpResponse);
  };
}
