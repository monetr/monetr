import Feature from '@monetr/docs/components/Feature';

import styles from './FeatureCard.module.scss';

interface FeatureCardProps {
  title: string;
  description: string;
  link?: string;
  linkText?: string;
  linkExternal?: boolean;
}

export default function FeatureCard(props: FeatureCardProps): JSX.Element {
  return (
    <Feature
      className={styles.root}
      description={<h2 className={styles.description}>{props.description}</h2>}
      link={props.link}
      linkExternal={props.linkExternal}
      linkText={props.linkText}
      title={<h1 className={styles.title}>{props.title}</h1>}
    />
  );
}
