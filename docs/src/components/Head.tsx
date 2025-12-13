import React, { Fragment } from 'react';
import { parse } from 'date-fns';

import realUrl from '@monetr/docs/components/utils/realUrl';

import { useRouter } from 'next/router';
import { useConfig } from 'nextra-theme-docs';

export default function Head(): JSX.Element {
  const { frontMatter, timestamp } = useConfig();
  const { asPath } = useRouter();

  const url = realUrl(asPath);
  let suffix = '';
  if (asPath !== '/') {
    suffix = '- monetr';
  }

  let title = 'monetr';
  if (frontMatter?.title?.length > 0) {
    title = `${frontMatter.title} ${suffix}`;
  }
  // Make a short title that excludes the suffix.
  const shortTitle = frontMatter?.title || title;

  let description =
    'Take control of your finances, paycheck by paycheck, with monetr. Put aside what you need, spend what you want, and confidently manage your money with ease. Always know you’ll have enough for your bills and what’s left to save or spend.';
  if (frontMatter?.description?.length > 0) {
    description = frontMatter.description;
  }

  let type = 'website';
  let authors: Array<string> = [];
  let published: Array<string> = [];
  let modified: Array<string> = [];
  if (frontMatter?.date) {
    type = 'article';
    authors = typeof frontMatter?.author === 'string' ? frontMatter.author.split(',') : [];
    published = [parse(frontMatter.date, 'yyyy/MM/dd', new Date()).toISOString()];
    modified = timestamp ? [new Date(timestamp).toISOString()] : [];
  }

  return (
    <React.Fragment>
      <meta content='width=device-width, initial-scale=1.0' name='viewport' />
      <title>{title}</title>
      <meta content={shortTitle} name='title' />
      <meta content={description} name='description' />
      <meta content={shortTitle} property='og:title' />
      <meta content={type} property='og:type' /> {/* TODO, make this change to article for blog posts */}
      <meta content={url} property='og:url' />
      <meta content={description} property='og:description' />
      {frontMatter?.ogImage && <meta content={realUrl(frontMatter.ogImage)} property='og:image' />}
      {authors.map(author => (
        <meta content={author} key={author} property='article:author' />
      ))}
      {published.map(timestamp => (
        <meta content={timestamp} key={timestamp} property='article:published_time' />
      ))}
      {modified.map(timestamp => (
        <meta content={timestamp} key={timestamp} property='article:modified_time' />
      ))}
      <meta content='en' httpEquiv='Content-Language' />
      <meta content='monetr.app' property='twitter:domain' />
      <meta content={url} property='twitter:url' />
      <meta content={shortTitle} name='twitter:title' />
      <meta content={description} name='twitter:description' />
      {frontMatter?.ogImage && (
        <Fragment>
          <meta content='summary_large_image' name='twitter:card' />
          <meta content={realUrl(frontMatter.ogImage)} name='twitter:image' />
        </Fragment>
      )}
      {process.env.NODE_ENV !== 'development' && (
        <script data-website-id='ccbdfaf9-683f-4487-b97f-5516e1353715' defer src='https://a.monetr.app/script.js' />
      )}
    </React.Fragment>
  );
}
