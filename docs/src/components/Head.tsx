import React from 'react';
import { useRouter } from 'next/router';
import { useConfig } from 'nextra-theme-docs';

import ChatwootIntegration from '@monetr/docs/components/ChatwootIntegration';

export default function Head(): JSX.Element {
  const { frontMatter } = useConfig();
  const { asPath } = useRouter();

  const url = `https://monetr.app${asPath}`;
  let suffix = '';
  if (asPath !== '/') {
    suffix = '- monetr';
  }

  let title = 'monetr';
  if (frontMatter?.title?.length > 0) {
    title = `${frontMatter.title} ${suffix}`;
  }

  let description = 'Take control of your finances, paycheck by paycheck, with monetr. Put aside what you need, spend what you want, and confidently manage your money with ease. Always know you’ll have enough for your bills and what’s left to save or spend.';
  if (frontMatter?.description?.length > 0) {
    description = frontMatter.description;
  }

  return (
    <React.Fragment>
      <meta name='viewport' content='width=device-width, initial-scale=1.0' />

      <title>{ title }</title>
      <meta property='og:title' content={ frontMatter.title || title } />
      <meta property='og:url' content={ url } />
      <meta name='title' content={ frontMatter.title || title } />
      <meta httpEquiv='Content-Language' content='en' />
      <meta property='og:description' content={ description } />
      <meta name='description' content={ description } />

      { process.env.NODE_ENV != 'development' && 
        <script defer src='https://a.monetr.app/script.js' data-website-id='ccbdfaf9-683f-4487-b97f-5516e1353715' /> 
      }
      <ChatwootIntegration />
    </React.Fragment>
  );
}
