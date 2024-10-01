import * as React from 'react';
import {
  Body,
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

interface PasswordChangedProps {
  baseUrl?: string;
  firstName?: string;
  lastName?: string;
  supportEmail?: string;
}

export const PasswordChanged = ({
  baseUrl = '{{ .BaseURL }}',
  firstName = '{{ .FirstName }}',
  lastName = '{{ .LastName }}',
  supportEmail = '{{ .SupportEmail }}',
}: PasswordChangedProps) => {
  const previewText = 'Your password has been updated';

  return (
    <Html>
      <Head />
      <Preview>{previewText}</Preview>
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
              Your password for <strong>monetr</strong> has been updated
            </Heading>
            <Text className='text-black text-sm leading-6'>
              Hello {firstName},
            </Text>
            <Text className='text-black text-sm leading-6'>
              If you did not initiate the change in your password please reach out to us immediately via our support
              email:{' '}
              <Link
                href={ `mailto:${ supportEmail }` }
                className='text-blue-600 no-underline'
              >
                { supportEmail }
              </Link>
            </Text>
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

PasswordChanged.PreviewProps = {
  baseUrl: 'https://my.monetr.dev',
  firstName: 'Elliot',
  lastName: 'Courant',
  supportEmail: 'support@monetr.local',
} as PasswordChangedProps;

export default PasswordChanged;


