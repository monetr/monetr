import { AxiosError } from 'axios';
import { useSnackbar, VariantType } from 'notistack';
import request from 'shared/util/request';

export default function useSendForgotPassword(): (email: string, ReCAPTCHA: string | null) => Promise<void> {
  const { enqueueSnackbar } = useSnackbar();
  return (email: string, ReCAPTCHA: string | null) => {
    return request().post('/authentication/forgot', {
      email,
      captcha: ReCAPTCHA,
    })
      .then(() => void enqueueSnackbar('Successfully sent password reset link.', {
        variant: 'success',
        disableWindowBlurListener: true,
      }))
      .catch((error: AxiosError) => {
        const message = error?.response?.data?.error || 'Failed to send password reset email.';
        let variant: VariantType = 'error';

        // Check to see if the status code is precondition required. If it is that means that the email address is valid
        // but is not verified. We are enforcing that an email address be verified before allowing any other actions to
        // be taken.
        if (error?.response?.status === 428) {
          // When we get this, change the notification to be a warning instead of an error.
          variant = 'warning';
        }

        enqueueSnackbar(message, {
          variant,
          disableWindowBlurListener: true,
        });
      });
  };
};
