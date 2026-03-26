import { Button, Heading, Hr, Link, Section, Text } from '../../components/email';
import EmailLayout from '../../components/EmailLayout';
import EmailLogo from '../../components/EmailLogo';
import styles from '../../styles/email.module.scss';

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
      <Heading className={styles.heading}>
        Reset your password for <strong>monetr</strong>
      </Heading>
      <Text className={styles.bodyText}>Hello {firstName},</Text>
      <Text className={styles.bodyText}>
        Below is the link you requested in order to change your login password.
      </Text>
      <Section className={styles.buttonSection}>
        <Button className={styles.button} href={resetUrl}>
          <Text className={styles.buttonText}>Reset password</Text>
        </Button>
      </Section>
      <Hr className={styles.hr} />
      <Text className={styles.footerText}>
        This message was intended for{' '}
        <span className={styles.footerName}>
          {firstName} {lastName}
        </span>
        . If you did not make this request or you are concerned about this communication please reach out to{' '}
        <Link className={styles.footerLink} href={`mailto:${supportEmail}`}>
          {supportEmail}
        </Link>
        .
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
