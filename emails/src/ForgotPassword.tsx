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
    <Html>
      <Head />
      <Preview>{previewText}</Preview>
      <Tailwind config={ tailwindConfig as any }>
        <Body className='bg-white my-auto mx-auto font-sans'>
          <Container className='border border-solid border-[#eaeaea] rounded my-10 mx-auto p-5 max-w-xl'>
            <Section className='mt-8'>
              <Img
                src={ `${baseUrl}/logo192.png` }
                width='64'
                height='64'
                alt='monetr'
                className='my-0 mx-auto'
              />
            </Section>
            <Heading className='text-black text-2xl font-normal text-center p-0 my-8 mx-0'>
              Reset your password for <strong>monetr</strong>
            </Heading>
            <Text className='text-black text-sm leading-6'>
              Hello {firstName},
            </Text>
            <Text className='text-black text-sm leading-6'>
              Below is the link you requested in order to change your login password.
            </Text>
            <Section className='text-center mt-9 mb-9'>
              <Button
                pX={ 20 }
                pY={ 12 }
                className='bg-purple-500 rounded-lg text-white text-xs font-semibold no-underline text-center'
                href={ resetUrl }
              >
                Reset password
              </Button>
            </Section>
            <Hr className='border border-solid border-[#eaeaea] my-6 mx-0 w-full' />
            <Text className='text-[#666666] text-xs leading-6'>
              This message was intended for{' '}
              <span className='text-black'>{firstName} {lastName}</span>.
              If you did not make this request or you are concerned about this communication please reach out to{' '}
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

ForgotPassword.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  supportEmail: 'support@monetr.local',
  resetUrl: 'https://monetr.local/test',
} as ForgotPasswordProps;

export default ForgotPassword;
