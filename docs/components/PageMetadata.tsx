import { parse } from 'date-fns';

import type { Author } from '@monetr/docs/components/hooks/useDocPages';
import realUrl from '@monetr/docs/components/utils/realUrl';

import { Head, usePage } from '@rspress/core/runtime';
import { Fragment } from 'react/jsx-runtime';

export default function PageMetadata(): React.JSX.Element {
  const { page } = usePage();
  const { frontmatter } = page;
  const url = realUrl(page.routePath);

  let type = 'website';
  let published: Array<string> = [];
  let modified: Array<string> = [];
  if (frontmatter?.date) {
    type = 'article';
    published = [parse(frontmatter?.date as string, 'yyyy/MM/dd', new Date()).toISOString()];
    modified = page.lastUpdatedTime ? [new Date(page.lastUpdatedTime).toISOString()] : [];
  }
  const authors = Array.isArray(frontmatter?.authors) ? (frontmatter?.authors as Array<Author>) : [];

  return (
    <Head>
      {/* opengraph things */}
      <meta content={type} property='og:type' />
      <meta content={url} property='og:url' />
      <meta content={page.title} name='title' />
      <meta content={page.title} name='og:title' /> {/* TODO This doesn't work! */}
      {page.description ? (
        <Fragment>
          <meta content={page.description} name='description' />
          <meta content={page.description} name='og:description' />
        </Fragment>
      ) : null}
      {frontmatter?.ogImage ? <meta content={realUrl(frontmatter.ogImage as string)} property='og:image' /> : null}
      {/* misc? */}
      <meta content={page.lang} httpEquiv='Content-Language' />
      {/* Article stuff */}
      {authors.map(author => (
        <meta content={author.name} key={author.name} property='article:author' />
      ))}
      {published.map(timestamp => (
        <meta content={timestamp} key={timestamp} property='article:published_time' />
      ))}
      {modified.map(timestamp => (
        <meta content={timestamp} key={timestamp} property='article:modified_time' />
      ))}
      {/* Twitter bullshit */}
      <meta content='monetr.app' property='twitter:domain' />
      <meta content={url} property='twitter:url' />
      <meta content={page.frontmatter?.title} name='twitter:title' />
      {/* TODO Doesn't work the same as the HeadTags description */}
      <meta content={page.frontmatter?.description} name='twitter:description' />{' '}
      {frontmatter?.ogImage ? (
        <Fragment>
          <meta content='summary_large_image' name='twitter:card' />
          <meta content={realUrl(frontmatter?.ogImage as string)} name='twitter:image' />
        </Fragment>
      ) : null}
    </Head>
  );
}
