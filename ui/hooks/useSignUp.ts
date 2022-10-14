import User from 'models/User';
import request from 'util/request';

export interface SignUpArguments {
  agree: boolean;
  betaCode: string | null;
  captcha: string | null;
  email: string;
  firstName: string;
  lastName: string;
  password: string;
  timezone: string;
}

export interface SignUpResponse {
  isActive: boolean;
  message: string | null;
  nextUrl: string | null;
  requireVerification: boolean | null;
  user: User | null;
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
