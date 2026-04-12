import EmailLayout from '@monetr/emails/components/EmailLayout';
import EmailLogo from '@monetr/emails/components/EmailLogo';
import Heading from '@monetr/emails/components/Heading';
import Hr from '@monetr/emails/components/Hr';
import Link from '@monetr/emails/components/Link';
import Typography from '@monetr/emails/components/Typography';

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
      <Heading>
        Your trial for <strong>monetr</strong> is about to expire
      </Heading>
      <Typography>Hello {firstName},</Typography>
      <Typography>
        We are just letting you know that your trial is about to expire. Don't worry, if you don't want to continue
        using monetr then no action is required on your part. If you would like to continue using monetr though, you
        will need to setup a subscription the next time you login.
      </Typography>
      <Typography>
        Your trial will expire in about <strong>{trialExpirationWindow}</strong> on{' '}
        <strong>{trialExpirationDate}</strong>.
      </Typography>
      <Typography>Thank you for trying out monetr!</Typography>
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

TrialAboutToExpire.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  trialExpirationDate: 'Monday October 1st, 2024',
  trialExpirationWindow: '3 days',
  supportEmail: 'support@monetr.local',
} as TrialAboutToExpireProps;

export default TrialAboutToExpire;
