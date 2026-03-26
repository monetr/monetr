import { Heading, Hr, Link, Text } from '../../components/email';
import EmailLayout from '../../components/EmailLayout';
import EmailLogo from '../../components/EmailLogo';
import styles from '../../styles/email.module.scss';

interface TrialAboutToExpireProps {
  baseUrl?: string;
  firstName?: string;
  lastName?: string;
  trialExpirationDate?: string;
  trialExpirationWindow?: string;
  supportEmail?: string;
}

export const TrialAboutToExpire = ({
  baseUrl = '{{ .BaseURL }}',
  firstName = '{{ .FirstName }}',
  lastName = '{{ .LastName }}',
  trialExpirationDate = '{{ .TrialExpirationDate }}',
  trialExpirationWindow = '{{ .TrialExpirationWindow }}',
  supportEmail = '{{ .SupportEmail }}',
}: TrialAboutToExpireProps) => {
  const previewText = 'Your monetr trial is about to expire';
  return (
    <EmailLayout previewText={previewText}>
      <EmailLogo baseUrl={baseUrl} />
      <Heading className={styles.heading}>
        Your trial for <strong>monetr</strong> is about to expire
      </Heading>
      <Text className={styles.bodyText}>Hello {firstName},</Text>
      <Text className={styles.bodyText}>
        We are just letting you know that your trial is about to expire. Don't worry, if you don't want to continue
        using monetr then no action is required on your part. If you would like to continue using monetr though, you
        will need to setup a subscription the next time you login.
      </Text>
      <Text className={styles.bodyText}>
        Your trial will expire in about <strong>{trialExpirationWindow}</strong> on{' '}
        <strong>{trialExpirationDate}</strong>.
      </Text>
      <Text className={styles.bodyText}>Thank you for trying out monetr!</Text>
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

TrialAboutToExpire.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  trialExpirationDate: 'Monday October 1st, 2024',
  trialExpirationWindow: '3 days',
  supportEmail: 'support@monetr.local',
} as TrialAboutToExpireProps;

export default TrialAboutToExpire;
