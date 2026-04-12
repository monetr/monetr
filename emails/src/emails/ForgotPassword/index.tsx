import EmailLayout from '@monetr/emails/components/EmailLayout';
import EmailLogo from '@monetr/emails/components/EmailLogo';
import { Button } from '@monetr/emails/components/email/Button';
import { Heading } from '@monetr/emails/components/email/Heading';
import { Hr } from '@monetr/emails/components/email/Hr';
import { Link } from '@monetr/emails/components/email/Link';
import { Typography } from '@monetr/emails/components/email/Typography';

interface ForgotPasswordProps {
  baseUrl?: string;
  firstName?: string;
  lastName?: string;
  supportEmail?: string;
  resetUrl?: string;
}

export const ForgotPassword = ({
  baseUrl = '{{ .BaseURL }}',
  firstName = '{{ .FirstName }}',
  lastName = '{{ .LastName }}',
  supportEmail = '{{ .SupportEmail }}',
  resetUrl = '{{ .ResetURL }}',
}: ForgotPasswordProps) => {
  const previewText = 'Reset your password for monetr';

  return (
    <EmailLayout previewText={previewText}>
      <EmailLogo baseUrl={baseUrl} />
      <Heading>
        Reset your password for <strong>monetr</strong>
      </Heading>
      <Typography>Hello {firstName},</Typography>
      <Typography>Below is the link you requested in order to change your login password.</Typography>
      <Button href={resetUrl}>Reset password</Button>
      <Hr />
      <Typography variant='footer'>
        This message was intended for{' '}
        <strong>
          {firstName} {lastName}
        </strong>
        . If you did not make this request or you are concerned about this communication please reach out to{' '}
        <Link href={`mailto:${supportEmail}`}>{supportEmail}</Link>.
      </Typography>
    </EmailLayout>
  );
};

ForgotPassword.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  supportEmail: 'support@monetr.local',
  resetUrl: 'https://monetr.local/test',
} as ForgotPasswordProps;

export default ForgotPassword;
