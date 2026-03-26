import EmailLayout from '../../components/EmailLayout';
import EmailLogo from '../../components/EmailLogo';
import { Button, Heading, Hr, Link, Section, Text } from '../../components/email';
import styles from '../../styles/email.module.scss';

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
      <Heading className={styles.heading}>
        Verify your email address for <strong>monetr</strong>
      </Heading>
      <Text className={styles.bodyText}>Hello {firstName},</Text>
      <Text className={styles.bodyText}>
        Thank you for signing up for monetr, in order to use your account we ask that you verify your email address.
      </Text>
      <Section className={styles.buttonSection}>
        <Button className={styles.button} href={inviteLink}>
          <Text className={styles.buttonText}>Verify email address</Text>
        </Button>
      </Section>
      <Hr className={styles.hr} />
      <Text className={styles.footerText}>
        This message was intended for{' '}
        <span className={styles.footerName}>
          {firstName} {lastName}
        </span>
        . If you did not sign up for <strong>monetr</strong>, you can ignore this email. If you are concerned about this
        communication please reach out to{' '}
        <Link className={styles.footerLink} href={`mailto:${supportEmail}`}>
          {supportEmail}
        </Link>
        .
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
