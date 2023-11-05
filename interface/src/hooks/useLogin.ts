import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';

import request from '@monetr/interface/util/request';

export interface LoginArguments {
  email: string;
  password: string;
  captcha?: string;
  totp?: string;
}

export default function useLogin(): (loginArgs: LoginArguments) => Promise<void> {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return async (loginArgs: LoginArguments): Promise<void> => {
    return request().post('/authentication/login', loginArgs)
      .then(async result => {
        // Then bootstrap the authentication, once it's bootstrapped we want to consider the `nextUrl` field from the
        // login response above. If the nextUrl is present, then we want to navigate the user to that path. If it is not
        // present then we can direct the user to the root path.
        return queryClient.invalidateQueries(['/users/me'])
          .then(() => navigate(result?.data?.nextUrl || '/'));
      })
      .catch(error => {
        // More important than the message though is the status of the response. If the status code was 428 then that
        // means the credentials are valid, but the user has not verified their email yet. If this is the case we want
        // to redirect them to the resend email verification page and autofill that user's email address.
        switch (error?.response?.status) {
          case 428: // The user needs to take some action before they can be fully authenticated.
            switch (error?.response?.data?.code) {
              case 'PASSWORD_CHANGE_REQUIRED':
                return navigate('/password/reset', {
                  state: {
                    'message': 'You are required to change your password before authenticating.',
                    'token': error?.response?.data?.resetToken,
                  },
                });
              case 'MFA_REQUIRED':
                return navigate('/login/mfa', {
                  state: {
                    'emailAddress': loginArgs.email,
                    'password': loginArgs.password,
                    // TODO ReCAPTCHA?
                  },
                });
              case 'EMAIL_NOT_VERIFIED':
                return navigate('/verify/email/resend', {
                  state: {
                    'emailAddress': loginArgs.email,
                  },
                });
              default:
                throw error;
            }
          case 403: // Invalid login.
            throw error;
        }

        throw error;
      });
  };
}
