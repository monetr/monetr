import { useSnackbar } from 'notistack';
import { useNavigate } from 'react-router-dom';

import type { ApiError } from '@monetr/interface/api/client';
import request, { type APIError } from '@monetr/interface/util/request';

export default function useResetPassword(): (newPassword: string, token: string) => Promise<void> {
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();
  return async (newPassword: string, token: string) => {
    return request({
      method: 'POST',
      url: '/api/authentication/reset',
      data: {
        token,
        password: newPassword,
      },
    })
      .then(() => {
        enqueueSnackbar('Password has been reset, please login with your new credentials.', {
          variant: 'success',
          disableWindowBlurListener: true,
        });
        navigate('/login');
      })
      .catch((error: ApiError<APIError>) => {
        const message = error.response.data.error || 'Failed to reset password.';
        enqueueSnackbar(message, {
          variant: 'error',
          disableWindowBlurListener: true,
        });

        throw error;
      });
  };
}
