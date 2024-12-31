import '@fontsource-variable/inter';

import type { ReactElement } from 'react';
import type { AppProps } from 'next/app';

import QueryClientWrapper from '@monetr/docs/components/QueryClientWrapper';

import '@monetr/docs/styles/globals.scss';

export default function App({ Component, pageProps }: AppProps): ReactElement {
  return (
    <QueryClientWrapper>
      <Component { ...pageProps } />
    </QueryClientWrapper>
  );
}
