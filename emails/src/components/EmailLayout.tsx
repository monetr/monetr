import type React from 'react';

import styles from './EmailLayout.module.scss';
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
      <Body className={styles.body}>
        <Container className={styles.container}>{props.children}</Container>
      </Body>
    </Html>
  );
}
