import User from 'models/User';
import { useNavigate } from 'react-router-dom';
import useBootstrapLogin from 'shared/authentication/actions/bootstrapLogin';
import request from 'shared/util/request';

export interface LoginArguments {
  email: string;
  password: string;
  captcha: string | null;
}

export default function useLogin(): (loginArgs: LoginArguments) => Promise<void> {
  const navigate = useNavigate();
  const bootstrapLogin = useBootstrapLogin();

  return (loginArgs: LoginArguments): Promise<void> => {
    return request().post('/authentication/login', loginArgs)
      .then(result => {
        // Then bootstrap the authentication, once it's bootstrapped we want to consider the `nextUrl` field from the
        // login response above. If the nextUrl is present, then we want to navigate the user to that path. If it is not
        // present then we can direct the user to the root path.
        return bootstrapLogin().then(() => navigate(result?.data?.nextUrl || '/'));
      })
      .catch(error => {
        // If there was an error logging in, then establish the reason why from the error message. But if that message
        // is not present in the response then just use a default "failed to authenticate" message.
        const errorMessage = error?.response?.data?.error || 'Failed to authenticate.';

        // More important than the message though is the status of the response. If the status code was 428 then that
        // means the credentials are valid, but the user has not verified their email yet. If this is the case we want
        // to redirect them to the resend email verification page and autofill that user's email address.
        switch (error?.response?.status) {
          case 428: // Email not verified.
            return navigate('/verify/email/resend', {
              state: {
                'emailAddress': loginArgs.email,
              }
            });
          case 403: // Invalid login.
            throw error;
        }

        throw error;
      });
  }
}