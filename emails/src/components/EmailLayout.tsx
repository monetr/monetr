import type React from 'react';
import { Body, Container, Head, Html, Preview } from './email';

export interface EmailLayoutProps {
  previewText: string;
  children: React.ReactNode;
}

export default function EmailLayout(props: EmailLayoutProps): JSX.Element {
  return (
    <Html>
      <Head />
      <Preview>{props.previewText}</Preview>
      <Body className='bg-white my-auto mx-auto font-sans'>
        <Container className='border border-solid border-gray-200 rounded my-10 mx-auto p-5 max-w-xl'>
          {props.children}
        </Container>
      </Body>
    </Html>
  );
}
