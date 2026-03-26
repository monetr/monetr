import { Button, Heading, Hr, Link, Section, Text } from '../../components/email';
import EmailLayout from '../../components/EmailLayout';
import EmailLogo from '../../components/EmailLogo';
import styles from '../../styles/email.module.scss';

interface PlaidDisconnectedProps {
  baseUrl?: string;
  firstName?: string;
  lastName?: string;
  linkName?: string;
  linkURL?: string;
  supportEmail?: string;
}

export const PlaidDisconnected = ({
  baseUrl = '{{ .BaseURL }}',
  firstName = '{{ .FirstName }}',
  lastName = '{{ .LastName }}',
  linkName = '{{ .LinkName }}',
  linkURL = '{{ .LinkURL }}',
  supportEmail = '{{ .SupportEmail }}',
}: PlaidDisconnectedProps) => {
  const previewText = 'One of your linked accounts has been disconnected';
  return (
    <EmailLayout previewText={previewText}>
      <EmailLogo baseUrl={baseUrl} />
      <Heading className={styles.heading}>
        One of your linked accounts has been disconnected
      </Heading>
      <Text className={styles.bodyText}>Hello {firstName},</Text>
      <Text className={styles.bodyText}>
        Your <strong>{linkName}</strong> account connected via Plaid needs to be reauthenticated. This account will not
        receive automatic updates until the link has been updated.
      </Text>
      <Section className={styles.buttonSection}>
        <Button className={styles.button} href={linkURL}>
          <Text className={styles.buttonText}>Reconnect {linkName}</Text>
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

PlaidDisconnected.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  linkName: 'Navy Federal Credit Union',
  linkURL: 'https://my.monetr.dev/bank_accounts/bac_123abc',
  supportEmail: 'support@monetr.local',
} as PlaidDisconnectedProps;

export default PlaidDisconnected;
