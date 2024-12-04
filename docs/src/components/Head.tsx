import React from 'react';
import { useRouter } from 'next/router';
import { useConfig } from 'nextra-theme-docs';

import ChatwootIntegration from '@monetr/docs/components/ChatwootIntegration';

export default function Head(): JSX.Element {
  const { frontMatter } = useConfig();
  const { asPath } = useRouter();

  let suffix = '';

  if (asPath !== '/') {
    suffix = '- monetr';
  }

  let title = 'monetr';
  if (frontMatter?.title) {
    title = `${frontMatter.title} ${suffix}`;
  }

  return (
    <React.Fragment>
      <meta name='viewport' content='width=device-width, initial-scale=1.0' />

      <title>{ title }</title>
      <meta property='og:title' content={ frontMatter.title || title } />
      <meta name='title' content={ frontMatter.title || title } />

      <meta property='og:description' content={ frontMatter.description || 'Take control of your finances, paycheck by paycheck, with monetr. Put aside what you need, spend what you want, and confidently manage your money with ease. Always know you’ll have enough for your bills and what’s left to save or spend.' } />
      <meta name='description' content={ frontMatter.description || 'monetr' } />

      { process.env.NODE_ENV != 'development' && 
        <script defer src='https://a.monetr.app/script.js' data-website-id='ccbdfaf9-683f-4487-b97f-5516e1353715' /> 
      }
      <ChatwootIntegration />
    </React.Fragment>
  );
}
