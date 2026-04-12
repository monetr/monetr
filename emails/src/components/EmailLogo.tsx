import Img from '@monetr/emails/components/Img';
import Section from '@monetr/emails/components/Section';

import styles from './EmailLogo.module.scss';

export interface EmailLogoProps {
  baseUrl: string;
}

export default function EmailLogo(props: EmailLogoProps): JSX.Element {
  return (
    <Section className={styles.section}>
      <Img
        alt='monetr'
        className={styles.logo}
        height='64'
        src={`${props.baseUrl}/assets/resources/transparent-128.png`}
        width='64'
      />
    </Section>
  );
}
