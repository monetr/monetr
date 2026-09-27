import EmailLayout from '@monetr/emails/components/EmailLayout';
import EmailLogo from '@monetr/emails/components/EmailLogo';
import Heading from '@monetr/emails/components/Heading';
import Hr from '@monetr/emails/components/Hr';
import Link from '@monetr/emails/components/Link';
import Typography from '@monetr/emails/components/Typography';

interface ApiKeyCreatedProps {
  baseUrl?: string;
  firstName?: string;
  lastName?: string;
  keyName?: string;
  createdByName?: string;
  createdByEmail?: string;
  supportEmail?: string;
}

export const ApiKeyCreated = ({
  baseUrl = '{{ .BaseURL }}',
  firstName = '{{ .FirstName }}',
  lastName = '{{ .LastName }}',
  keyName = '{{ .KeyName }}',
  createdByName = '{{ .CreatedByName }}',
  createdByEmail = '{{ .CreatedByEmail }}',
  supportEmail = '{{ .SupportEmail }}',
}: ApiKeyCreatedProps) => {
  const previewText = 'A new API key was created for your account';
  return (
    <EmailLayout previewText={previewText}>
      <EmailLogo baseUrl={baseUrl} />
      <Heading>
        A new API key was created for your <strong>monetr</strong> account
      </Heading>
      <Typography>Hello {firstName},</Typography>
      <Typography>
        An API key named <strong>{keyName}</strong> was just created for your account by{' '}
        <strong>{createdByName}</strong> ({createdByEmail}). API keys can be used to access your account's data.
      </Typography>
      <Typography>
        If you did not expect this API key to be created please revoke it from your account settings and reach out to us
        immediately via our support email: <Link href={`mailto:${supportEmail}`}>{supportEmail}</Link>
      </Typography>
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

ApiKeyCreated.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  keyName: 'My First Key',
  createdByName: 'Elliot Courant',
  createdByEmail: 'elliot@monetr.local',
  supportEmail: 'support@monetr.local',
} as ApiKeyCreatedProps;

export default ApiKeyCreated;
