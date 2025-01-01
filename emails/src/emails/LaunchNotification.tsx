import * as React from 'react';
import {
    Body,
  Button,
  Container,
  Head,
  Heading,
  Hr,
  Html,
  Link,
  Preview,
  Section,
  Tailwind,
  Text,
} from '@react-email/components';

import EmailLayout from '../components/EmailLayout';
import EmailLogo from '../components/EmailLogo';
// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import tailwindConfig from '../../tailwind.config.ts';

interface LaunchNotificationProps {
  baseUrl?: string;
  supportEmail?: string;
}

export const LaunchNotification = ({
  baseUrl = '{{ .BaseURL }}',
  supportEmail = '{{ .SupportEmail }}',
}: LaunchNotificationProps) => {
  const previewText = 'monetr is now live!';

  return (
    <Html>
      <Head />
      <Preview>{previewText}</Preview>
      <Tailwind config={tailwindConfig as any}>
        <Body className='bg-white my-auto mx-auto font-sans'>
          <Container className='border-0 my-10 mx-auto p-5 max-w-xl'>
            <EmailLogo baseUrl={baseUrl} />
            <Heading className='text-black text-2xl font-normal text-center p-0 my-8 mx-0'>
              monetr is now live!
            </Heading>
            <Text className='text-black text-sm leading-6'>
              Thank you for signing up for monetr's launch notification. monetr is now open for registrations. You will be
              required to verify your email as part of the sign up process but you will not be prompted for a payment method
              until the conclusion of your trial. We will notify you a few days before your trial ends.
            </Text>
            <Section className='text-center mt-9 mb-9 border-0'>
              <Button
                className='bg-purple-500 rounded-lg text-white text-sm font-semibold no-underline text-center'
                href={ 'https://my.monetr.app/register' }
              >
                <Text className='text-sm text-white m-2'>
                  Sign Up for monetr
                </Text>
              </Button>
            </Section>
            <Text className='text-black text-sm leading-6'>
              Thank you again and I really hope you enjoy monetr, if you have any feedback at all please reach out!
            </Text>
            <Text className='text-black text-sm leading-6'>
              -Elliot Courant
            </Text>
            <Hr className='border border-solid border-gray-200 my-6 mx-0 w-full' />
            <Text className='text-gray-500 text-xs leading-6'>
              If you did not make this request or you are concerned about this communication please reach out to{' '}
              <Link
                href={`mailto:${supportEmail}`}
                className='text-blue-600 no-underline'
              >
                {supportEmail}
              </Link>.

              You may also unsubscribe from these notifications below.
            </Text>
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
};

LaunchNotification.PreviewProps = {
  baseUrl: 'https://my.monetr.app',
  supportEmail: 'support@monetr.app',
} as LaunchNotificationProps;

export default LaunchNotification;
