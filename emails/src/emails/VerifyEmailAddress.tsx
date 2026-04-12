import Button from '@monetr/emails/components/Button';
import EmailLayout from '@monetr/emails/components/EmailLayout';
import EmailLogo from '@monetr/emails/components/EmailLogo';
import Heading from '@monetr/emails/components/Heading';
import Hr from '@monetr/emails/components/Hr';
import Link from '@monetr/emails/components/Link';
import Typography from '@monetr/emails/components/Typography';

interface VerifyEmailProps {
  baseUrl?: string;
  firstName?: string;
  lastName?: string;
  supportEmail?: string;
  verifyLink?: string;
}

export const VerifyEmailAddress = ({
  baseUrl = '{{ .BaseURL }}',
  firstName = '{{ .FirstName }}',
  lastName = '{{ .LastName }}',
  supportEmail = '{{ .SupportEmail }}',
  verifyLink: inviteLink = '{{ .VerifyURL }}',
}: VerifyEmailProps) => {
  const previewText = 'Verify your email address for monetr';
  return (
    <EmailLayout previewText={previewText}>
      <EmailLogo baseUrl={baseUrl} />
      <Heading>
        Verify your email address for <strong>monetr</strong>
      </Heading>
      <Typography>Hello {firstName},</Typography>
      <Typography>
        Thank you for signing up for monetr, in order to use your account we ask that you verify your email address.
      </Typography>
      <Button href={inviteLink}>Verify email address</Button>
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

VerifyEmailAddress.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  supportEmail: 'support@monetr.local',
  verifyLink: 'https://monetr.local/test',
} as VerifyEmailProps;

export default VerifyEmailAddress;
