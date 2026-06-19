import { useLocation } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import request, { type APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

export default function useResetPassword(): (newPassword: string, token: string) => Promise<void> {
  const { enqueueSnackbar } = useSnackbar();
  const [, navigate] = useLocation();
  return async (newPassword: string, token: string) => {
    return await request({
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
