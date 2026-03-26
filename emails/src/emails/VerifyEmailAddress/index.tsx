import { Button, Heading, Hr, Link, Text } from '../../components/email';
import EmailLayout from '../../components/EmailLayout';
import EmailLogo from '../../components/EmailLogo';

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
      <Text>Hello {firstName},</Text>
      <Text>
        Thank you for signing up for monetr, in order to use your account we ask that you verify your email address.
      </Text>
      <Button href={inviteLink}>Verify email address</Button>
      <Hr />
      <Text variant='footer'>
        This message was intended for{' '}
        <span style={{ color: '#000' }}>
          {firstName} {lastName}
        </span>
        . If you did not sign up for <strong>monetr</strong>, you can ignore this email. If you are concerned about this
        communication please reach out to{' '}
        <Link href={`mailto:${supportEmail}`}>{supportEmail}</Link>.
      </Text>
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
