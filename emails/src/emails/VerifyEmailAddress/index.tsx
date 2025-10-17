import * as React from 'react';
import { Button, Heading, Hr, Link, Section, Text } from '@react-email/components';

import EmailLayout from '../../components/EmailLayout';
import EmailLogo from '../../components/EmailLogo';

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
      <Heading className='text-black text-2xl font-normal text-center p-0 my-8 mx-0'>
        Verify your email address for <strong>monetr</strong>
      </Heading>
      <Text className='text-black text-sm leading-6'>Hello {firstName},</Text>
      <Text className='text-black text-sm leading-6'>
        Thank you for signing up for monetr, in order to use your account we ask that you verify your email address.
      </Text>
      <Section className='text-center mt-9 mb-9'>
        <Button
          className='bg-purple-500 rounded-lg text-white text-sm font-semibold no-underline text-center'
          href={inviteLink}
        >
          <Text className='text-sm text-white m-2'>Verify email address</Text>
        </Button>
      </Section>
      <Hr className='border border-solid border-gray-200 my-6 mx-0 w-full' />
      <Text className='text-gray-500 text-xs leading-6'>
        This message was intended for{' '}
        <span className='text-black'>
          {firstName} {lastName}
        </span>
        . If you did not sign up for <strong>monetr</strong>, you can ignore this email. If you are concerned about this
        communication please reach out to{' '}
        <Link href={`mailto:${supportEmail}`} className='text-blue-600 no-underline'>
          {supportEmail}
        </Link>
        .
      </Text>
    </EmailLayout>
  );
};

VerifyEmailAddress.PreviewProps = {
  baseUrl: 'https://my.monetr.dev', // '{{ .BaseURL }}',
  firstName: 'Elliot',
  lastName: 'Courant',
  supportEmail: 'support@monetr.local',
  verifyLink: 'https://monetr.local/test',
} as VerifyEmailProps;

export default VerifyEmailAddress;
