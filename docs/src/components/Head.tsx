import React from 'react';
import { useRouter } from 'next/router';
import { useConfig } from 'nextra-theme-docs';

import ChatwootIntegration from '@monetr/docs/components/ChatwootIntegration';

export default function Head(): JSX.Element {
  const { frontMatter } = useConfig();
  const { asPath } = useRouter();

  let title = 'monetr';
  if (asPath !== '/') {
    title = Boolean(frontMatter?.title) ? `${frontMatter.title} - monetr` : 'monetr';
  }

  return (
    <React.Fragment>
      <meta name='viewport' content='width=device-width, initial-scale=1.0' />

      <title>{ title }</title>
      <meta property='og:title' content={ frontMatter.title || 'monetr' } />
      <meta name='title' content={ frontMatter.title || 'monetr' } />

      <meta property='og:description' content={ frontMatter.description || 'Transparent financial planning' } />
      <meta name='description' content={ frontMatter.title || 'monetr' } />

      { process.env.NODE_ENV != 'development' && 
        <script defer src='https://a.monetr.app/script.js' data-website-id='ccbdfaf9-683f-4487-b97f-5516e1353715' /> 
      }
      <ChatwootIntegration />
    </React.Fragment>
  );
}
