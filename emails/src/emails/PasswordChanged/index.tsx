import { Heading, Hr, Link, Text } from '../../components/email';
import EmailLayout from '../../components/EmailLayout';
import EmailLogo from '../../components/EmailLogo';

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
      <Text>Hello {firstName},</Text>
      <Text>
        If you did not initiate the change in your password please reach out to us immediately via our support email:{' '}
        <Link href={`mailto:${supportEmail}`}>{supportEmail}</Link>
      </Text>
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

PasswordChanged.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  supportEmail: 'support@monetr.local',
} as PasswordChangedProps;

export default PasswordChanged;
