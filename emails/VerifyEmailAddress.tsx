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

import tailwindConfig from '../tailwind.config.cjs';

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
    <Html>
      <Head />
      <Preview>{previewText}</Preview>
      <Tailwind config={ tailwindConfig as any }>
        <Body className="bg-white my-auto mx-auto font-sans">
          <Container className="border border-solid border-[#eaeaea] rounded my-10 mx-auto p-5 max-w-xl">
            <Section className="mt-8">
              <Img
                src={ `${baseUrl}/logo192.png` }
                width="64"
                height="64"
                alt="monetr"
                className="my-0 mx-auto"
              />
            </Section>
            <Heading className="text-black text-2xl font-normal text-center p-0 my-8 mx-0">
              Verify your email address for <strong>monetr</strong>
            </Heading>
            <Text className="text-black text-sm leading-6">
              Hello {firstName},
            </Text>
            <Text className="text-black text-sm leading-6">
              Thank you for signing up for monetr, in order to use your account we ask that you verify your email
              address.
            </Text>
            <Section className="text-center mt-9 mb-9">
              <Button
                pX={ 20 }
                pY={ 12 }
                className="bg-purple-500 rounded text-white text-xs font-semibold no-underline text-center"
                href={ inviteLink }
              >
                Verify email address
              </Button>
            </Section>
            <Hr className="border border-solid border-[#eaeaea] my-6 mx-0 w-full" />
            <Text className="text-[#666666] text-xs leading-6">
              This message was intended for{' '}
              <span className="text-black">{firstName} {lastName}</span>.
              If you did not sign up for <strong>monetr</strong>, you can ignore this email. If you are concerned about
              this communication please reach out to{' '}
              <Link
                href={ `mailto:${ supportEmail }` }
                className="text-blue-600 no-underline"
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

VerifyEmailAddress.PreviewProps = {
  baseUrl: 'https://my.monetr.dev', // '{{ .BaseURL }}',
  firstName: 'Elliot',
  lastName: 'Courant',
  supportEmail: 'support@monetr.local',
  verifyLink: 'https://monetr.local/test',
} as VerifyEmailProps;

export default VerifyEmailAddress;
