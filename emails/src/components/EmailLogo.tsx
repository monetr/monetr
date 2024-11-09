import React from 'react';
import { Img, Section } from '@react-email/components';

export interface EmailLogoProps {
  baseUrl: string;
}

export default function EmailLogo(props: EmailLogoProps): JSX.Element {
  return (
    <Section className='mt-8'>
      <Img
        src={ `${props.baseUrl}/assets/resources/android-chrome-192-192.png ` }
        width='64'
        height='64'
        alt='monetr'
        className='my-0 mx-auto'
      />
    </Section>
  );
}
