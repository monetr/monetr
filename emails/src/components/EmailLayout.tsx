import type React from 'react';

import Body from '@monetr/emails/components/Body';
import Container from '@monetr/emails/components/Container';
import Head from '@monetr/emails/components/Head';
import Html from '@monetr/emails/components/Html';
import Preview from '@monetr/emails/components/Preview';

import styles from './EmailLayout.module.scss';

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
