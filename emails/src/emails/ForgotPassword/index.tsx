import { Button, Heading, Hr, Link, Section, Text } from '@react-email/components';

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
      <Heading className='text-black text-2xl font-normal text-center p-0 my-8 mx-0'>
        Reset your password for <strong>monetr</strong>
      </Heading>
      <Text className='text-black text-sm leading-6'>Hello {firstName},</Text>
      <Text className='text-black text-sm leading-6'>
        Below is the link you requested in order to change your login password.
      </Text>
      <Section className='text-center mt-9 mb-9'>
        <Button
          className='bg-purple-500 rounded-lg text-white text-sm font-semibold no-underline text-center'
          href={resetUrl}
        >
          <Text className='text-sm text-white m-2'>Reset password</Text>
        </Button>
      </Section>
      <Hr className='border border-solid border-gray-200 my-6 mx-0 w-full' />
      <Text className='text-gray-500 text-xs leading-6'>
        This message was intended for{' '}
        <span className='text-black'>
          {firstName} {lastName}
        </span>
        . If you did not make this request or you are concerned about this communication please reach out to{' '}
        <Link href={`mailto:${supportEmail}`} className='text-blue-600 no-underline'>
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
