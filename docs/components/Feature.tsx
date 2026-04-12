import type React from 'react';

import mergeClasses from '@monetr/docs/util/mergeClasses';

import styles from './Feature.module.scss';

import { Link } from '@rspress/core/theme-original';

interface FeatureProps {
  title: React.ReactNode;
  description?: React.ReactNode;
  className?: string;
  link?: string;
  linkText?: React.ReactNode;
  linkExternal?: boolean;
}

export default function Feature(props: FeatureProps): JSX.Element {
  return (
    <div className={mergeClasses(styles.root, props.className)}>
      <div className={styles.body}>
        {props.title}
        {props.description && props.description}
      </div>
      {props.link && (
        <Link
          className={styles.link}
          href={props.link}
          rel={props.linkExternal ? 'noreferrer' : undefined}
          target={props.linkExternal ? '_blank' : undefined}
        >
          {props.linkText ?? 'Learn More'}
        </Link>
      )}
    </div>
  );
}
