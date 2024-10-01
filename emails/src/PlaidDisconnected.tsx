import * as React from 'react';
import {
  Body,
  Button,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Img,
  Link,
  Preview,
  Section,
  Tailwind,
  Text,
} from '@react-email/components';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import tailwindConfig from '../tailwind.config.js';

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
    <Html>
      <Head />
      <Preview>{ previewText }</Preview>
      <Tailwind config={ tailwindConfig as any }>
        <Body className='bg-white my-auto mx-auto font-sans'>
          <Container className='border border-solid border-[#eaeaea] rounded my-10 mx-auto p-5 max-w-xl'>
            <Section className='mt-8'>
              <Img
                src={ `${baseUrl}/logo192transparent.png` }
                width='64'
                height='64'
                alt='monetr'
                className='my-0 mx-auto'
              />
            </Section>
            <Heading className='text-black text-2xl font-normal text-center p-0 my-8 mx-0'>
              One of your linked accounts has been disconnected
            </Heading>
            <Text className='text-black text-sm leading-6'>
              Hello {firstName},
            </Text>
            <Text className='text-black text-sm leading-6'>
              Your <strong>{ linkName }</strong> account connected via Plaid needs to be reauthenticated. This account
              will not receive automatic updates until the link has been updated.
            </Text>
            <Section className='text-center mt-9 mb-9'>
              <Button
                className='bg-purple-500 rounded-lg text-white text-sm font-semibold no-underline text-center'
                href={ linkURL }
              >
                <Text className='text-sm text-white m-2'>
                  Reconnect { linkName }
                </Text>
              </Button>
            </Section>
            <Hr className='border border-solid border-[#eaeaea] my-6 mx-0 w-full' />
            <Text className='text-[#666666] text-xs leading-6'>
              This message was intended for{' '}
              <span className='text-black'>{firstName} {lastName}</span>.
              If you did not sign up for <strong>monetr</strong>, you can ignore this email. If you are concerned about
              this communication please reach out to{' '}
              <Link
                href={ `mailto:${ supportEmail }` }
                className='text-blue-600 no-underline'
              >
                { supportEmail }
              </Link>.
            </Text>
          </Container>
        </Body>
      </Tailwind>
    </Html>
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


