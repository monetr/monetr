import * as React from 'react';
import {
  Heading,
  Hr,
  Link,
  Text,
} from '@react-email/components';

import EmailLayout from '../components/EmailLayout';
import EmailLogo from '../components/EmailLogo';

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
      <EmailLogo baseUrl={ baseUrl } />
      <Heading className='text-black text-2xl font-normal text-center p-0 my-8 mx-0'>
        Your trial for <strong>monetr</strong> is about to expire
      </Heading>
      <Text className='text-black text-sm leading-6'>
        Hello {firstName},
      </Text>
      <Text className='text-black text-sm leading-6'>
        We are just letting you know that your trial is about to expire. Don't worry, if you don't want to continue
        using monetr then no action is required on your part. If you would like to continue using monetr though, you
        will need to setup a subscription the next time you login.
      </Text>
      <Text className='text-black text-sm leading-6'>
        Your trial will expire in about <strong>{ trialExpirationWindow }</strong> on <strong>{ trialExpirationDate }</strong>.
      </Text>
      <Text className='text-black text-sm leading-6'>
        Thank you for trying out monetr!
      </Text>
      <Hr className='border border-solid border-gray-200 my-6 mx-0 w-full' />
      <Text className='text-gray-500 text-xs leading-6'>
        This message was intended for{' '}
        <span className='text-black'>{firstName} {lastName}</span>.
        If you did not sign up for <strong>monetr</strong>, you can ignore this email. If you are concerned about
        this communication please reach out to{' '}
        <Link
          href={`mailto:${supportEmail}`}
          className='text-blue-600 no-underline'
        >
          {supportEmail}
        </Link>.
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


