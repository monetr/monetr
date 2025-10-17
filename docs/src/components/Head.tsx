import React, { Fragment } from 'react';
import { parse } from 'date-fns';
import { useRouter } from 'next/router';
import { useConfig } from 'nextra-theme-docs';

import ChatwootIntegration from '@monetr/docs/components/ChatwootIntegration';
import realUrl from '@monetr/docs/components/utils/realUrl';

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
      <meta name='viewport' content='width=device-width, initial-scale=1.0' />
      <title>{title}</title>
      <meta name='title' content={shortTitle} />
      <meta name='description' content={description} />
      <meta property='og:title' content={shortTitle} />
      <meta property='og:type' content={type} /> {/* TODO, make this change to article for blog posts */}
      <meta property='og:url' content={url} />
      <meta property='og:description' content={description} />
      {frontMatter?.ogImage && <meta property='og:image' content={realUrl(frontMatter.ogImage)} />}
      {authors.map(author => (
        <meta key={author} property='article:author' content={author} />
      ))}
      {published.map(timestamp => (
        <meta key={timestamp} property='article:published_time' content={timestamp} />
      ))}
      {modified.map(timestamp => (
        <meta key={timestamp} property='article:modified_time' content={timestamp} />
      ))}
      <meta httpEquiv='Content-Language' content='en' />
      <meta property='twitter:domain' content='monetr.app' />
      <meta property='twitter:url' content={url} />
      <meta name='twitter:title' content={shortTitle} />
      <meta name='twitter:description' content={description} />
      {frontMatter?.ogImage && (
        <Fragment>
          <meta name='twitter:card' content='summary_large_image' />
          <meta name='twitter:image' content={realUrl(frontMatter.ogImage)} />
        </Fragment>
      )}
      {process.env.NODE_ENV != 'development' && (
        <script defer src='https://a.monetr.app/script.js' data-website-id='ccbdfaf9-683f-4487-b97f-5516e1353715' />
      )}
      <ChatwootIntegration />
    </React.Fragment>
  );
}
