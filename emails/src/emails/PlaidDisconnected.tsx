import EmailLayout from '@monetr/emails/components/EmailLayout';
import EmailLogo from '@monetr/emails/components/EmailLogo';
import Button from '@monetr/emails/components/Button';
import Heading from '@monetr/emails/components/Heading';
import Hr from '@monetr/emails/components/Hr';
import Link from '@monetr/emails/components/Link';
import Typography from '@monetr/emails/components/Typography';

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
      <Heading>One of your linked accounts has been disconnected</Heading>
      <Typography>Hello {firstName},</Typography>
      <Typography>
        Your <strong>{linkName}</strong> account connected via Plaid needs to be reauthenticated. This account will not
        receive automatic updates until the link has been updated.
      </Typography>
      <Button href={linkURL}>Reconnect {linkName}</Button>
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

PlaidDisconnected.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  linkName: 'Navy Federal Credit Union',
  linkURL: 'https://my.monetr.dev/bank_accounts/bac_123abc',
  supportEmail: 'support@monetr.local',
} as PlaidDisconnectedProps;

export default PlaidDisconnected;
