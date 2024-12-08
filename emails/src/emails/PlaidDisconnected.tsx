import * as React from 'react';
import {
  Button,
  Heading,
  Hr,
  Link,
  Section,
  Text,
} from '@react-email/components';

import EmailLayout from '../components/EmailLayout';
import EmailLogo from '../components/EmailLogo';

interface PlaidDisconnectedProps {
  baseUrl?: string;
  firstName?: string;
  lastName?: string;
  linkName?: string;
  linkURL?: string;
  supportEmail?: string;
  goTemplate?: boolean;
}

export const PlaidDisconnected = ({
  baseUrl = '{{ .BaseURL }}',
  firstName = '{{ .FirstName }}',
  lastName = '{{ .LastName }}',
  linkName = '{{ .LinkName }}',
  linkURL = '{{ .LinkURL }}',
  supportEmail = '{{ .SupportEmail }}',
  goTemplate = true,
}: PlaidDisconnectedProps) => {
  const previewText = 'One of your linked accounts has been disconnected';
  return (
    <EmailLayout previewText={previewText}>
      <EmailLogo baseUrl={ baseUrl } />
      <Heading className='text-black text-2xl font-normal text-center p-0 my-8 mx-0'>
        One of your linked accounts has been disconnected
      </Heading>
      <Text className='text-black text-sm leading-6'>
        Hello {firstName},
      </Text>
      <Text className='text-black text-sm leading-6'>
        Your <strong>{linkName}</strong> account connected via Plaid needs to be reauthenticated. This account
        will not receive automatic updates until the link has been updated.
      </Text>
      <Section className='text-center mt-9 mb-9'>
        <Button
          className='bg-purple-500 rounded-lg text-white text-sm font-semibold no-underline text-center'
          href={linkURL}
        >
          <Text className='text-sm text-white m-2'>
            Reconnect {linkName}
          </Text>
        </Button>
      </Section>
      <Hr className='border border-solid border-gray-200 my-6 mx-0 w-full' />
      <Text className='text-gray-500 text-xs leading-6'>
        This message was intended for{' '}
        <span className='text-black'>{firstName} {lastName}</span>.
        { goTemplate ? '{{ if .SupportEmail }}' : '' }
        If you did not sign up for <strong>monetr</strong>, you can ignore this email. If you are concerned about
        this communication please reach out to{' '}
        <Link
          href={`mailto:${supportEmail}`}
          className='text-blue-600 no-underline'
        >
          {supportEmail}
        </Link>.
        { goTemplate ? '{{ end }}' : '' }
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
  goTemplate: false,
} as PlaidDisconnectedProps;

export default PlaidDisconnected;


