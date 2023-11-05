import { useNavigate } from 'react-router-dom';
import { AxiosError } from 'axios';
import { useSnackbar } from 'notistack';

import request, { APIError } from '@monetr/interface/util/request';

export default function useResetPassword(): (newPassword: string, token: string) => Promise<void> {
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();
  return async (newPassword: string, token: string) => {
    return request().post('/authentication/reset', {
      token,
      password: newPassword,
    })
      .then(() => {
        enqueueSnackbar('Password has been reset, please login with your new credentials.', {
          variant: 'success',
          disableWindowBlurListener: true,
        });
        navigate('/login');
      })
      .catch((error: AxiosError<APIError>) => {
        const message = error.response.data.error || 'Failed to reset password.';
        enqueueSnackbar(message, {
          variant: 'error',
          disableWindowBlurListener: true,
        });

        throw error;
      });
  };
}
