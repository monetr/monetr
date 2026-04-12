import { Button, Heading, Hr, Link, Text } from '../../components/email';
import textStyles from '../../components/email/Text.module.scss';
import EmailLayout from '../../components/EmailLayout';
import EmailLogo from '../../components/EmailLogo';

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
      <Text>Hello {firstName},</Text>
      <Text>Below is the link you requested in order to change your login password.</Text>
      <Button href={resetUrl}>Reset password</Button>
      <Hr />
      <Text variant='footer'>
        This message was intended for{' '}
        <span className={textStyles.recipient}>
          {firstName} {lastName}
        </span>
        . If you did not make this request or you are concerned about this communication please reach out to{' '}
        <Link href={`mailto:${supportEmail}`}>{supportEmail}</Link>.
      </Text>
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
