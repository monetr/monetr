/*
  @license
  This code is adapted from https://github.com/langfuse/langfuse-docs

  MIT License

  Copyright (c) 2022 Finto Technologies

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in all
  copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
  SOFTWARE.
  @license-end
*/

import { format, parse } from 'date-fns';
import { ArrowRight } from 'lucide-react';

import GradientHeading from '@monetr/docs/components/GradientHeading/GradientHeading';
import useDocPages from '@monetr/docs/components/hooks/useDocPages';
import mergeClasses from '@monetr/docs/util/mergeClasses';

import styles from './BlogIndex.module.scss';

import { Link } from '@rspress/core/theme-original';

export default function BlogIndex(): JSX.Element {
  const { pages } = useDocPages();

  const blogPages = pages
    .filter(page => page.routePath.startsWith('/blog/') && page.routePath !== '/blog/' && page.routePath !== '/blog')
    .sort((a, b) => {
      const dateA = a.frontmatter?.date ? new Date(a.frontmatter.date as string).getTime() : 0;
      const dateB = b.frontmatter?.date ? new Date(b.frontmatter.date as string).getTime() : 0;
      return dateB - dateA;
    });

  return (
    <div className={mergeClasses(styles.root, 'm-view-height')}>
      <div className={styles.header}>
        <GradientHeading
          blurClassName={styles.titleBlur}
          foregroundClassName={styles.titleForeground}
          wrapperClassName={styles.titleWrapper}
        >
          Blog
        </GradientHeading>
      </div>
      <div className={mergeClasses(styles.list, 'm-view-width')}>
        {blogPages.map(page => (
          <Link className={styles.cardLink} href={page.routePath} key={page.routePath}>
            {(page.frontmatter?.ogImage as string) ? (
              <div className={styles.imageWrap}>
                <img
                  alt={(page.frontmatter?.title as string) ?? 'Blog post image'}
                  className={styles.image}
                  src={page.frontmatter.ogImage as string}
                />
              </div>
            ) : null}
            <h2 className={styles.titleRow}>
              {(page.frontmatter?.title as string) || page.title}
              {(page.frontmatter?.tag as string) ? (
                <span className={styles.titleTag}>{page.frontmatter.tag as string}</span>
              ) : null}
            </h2>
            <div className={styles.description}>
              {page.frontmatter?.description as string}
              &nbsp;
              <span className={styles.descriptionReadMore}>
                Read more <ArrowRight className={styles.descriptionReadMoreArrow} />
              </span>
            </div>
            <div className={styles.meta}>
              {(page.frontmatter?.date as string) ? (
                <span className={styles.metaItem}>
                  {format(parse(page.frontmatter.date as string, 'yyyy/MM/dd', new Date()), 'MMMM dd, yyyy')}
                </span>
              ) : null}
              {page?.authors?.map(author => (
                <span className={styles.metaItem} key={author.name}>
                  by {author.name}
                </span>
              ))}
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
