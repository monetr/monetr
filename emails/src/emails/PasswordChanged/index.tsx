import EmailLayout from '@monetr/emails/components/EmailLayout';
import EmailLogo from '@monetr/emails/components/EmailLogo';
import { Heading } from '@monetr/emails/components/email/Heading';
import { Hr } from '@monetr/emails/components/email/Hr';
import { Link } from '@monetr/emails/components/email/Link';
import { Typography } from '@monetr/emails/components/email/Typography';

interface PasswordChangedProps {
  baseUrl?: string;
  firstName?: string;
  lastName?: string;
  supportEmail?: string;
}

export const PasswordChanged = ({
  baseUrl = '{{ .BaseURL }}',
  firstName = '{{ .FirstName }}',
  lastName = '{{ .LastName }}',
  supportEmail = '{{ .SupportEmail }}',
}: PasswordChangedProps) => {
  const previewText = 'Your password has been updated';
  return (
    <EmailLayout previewText={previewText}>
      <EmailLogo baseUrl={baseUrl} />
      <Heading>
        Your password for <strong>monetr</strong> has been updated
      </Heading>
      <Typography>Hello {firstName},</Typography>
      <Typography>
        If you did not initiate the change in your password please reach out to us immediately via our support email:{' '}
        <Link href={`mailto:${supportEmail}`}>{supportEmail}</Link>
      </Typography>
      <Hr />
      <Typography variant='footer'>
        This message was intended for{' '}
        <strong>
          {firstName} {lastName}
        </strong>
        . If you did not sign up for <strong>monetr</strong>, you can ignore this email. If you are concerned about this
        communication please reach out to <Link href={`mailto:${supportEmail}`}>{supportEmail}</Link>.
      </Typography>
    </EmailLayout>
  );
};

PasswordChanged.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  supportEmail: 'support@monetr.local',
} as PasswordChangedProps;

export default PasswordChanged;
