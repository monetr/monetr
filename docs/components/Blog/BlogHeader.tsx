import { useMemo } from 'react';
import { format, parse } from 'date-fns';

import GradientHeading from '@monetr/docs/components/GradientHeading/GradientHeading';

import styles from './BlogHeader.module.scss';

import { useFrontmatter } from '@rspress/core/runtime';

export default function BlogHeader(): JSX.Element {
  const { frontmatter } = useFrontmatter();

  const date = useMemo(() => {
    const blogDate = (frontmatter as Record<string, string>)?.date;
    if (blogDate) {
      return format(parse(blogDate, 'yyyy/MM/dd', new Date()), 'MMMM dd, yyyy');
    }
  }, [frontmatter]);

  return (
    <div className={styles.root}>
      <span className={styles.date}>{date}</span>
      <GradientHeading
        blurClassName={styles.titleBlur}
        foregroundClassName={styles.titleForeground}
        wrapperClassName={styles.titleWrapper}
      >
        {frontmatter?.title as string}
      </GradientHeading>
      <p className={styles.description}>{frontmatter?.description as string}</p>
    </div>
  );
}
