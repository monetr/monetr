import React from 'react';
import { Location, MemoryRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { Queries, queries, render, RenderOptions, RenderResult } from '@testing-library/react';

import MQueryClient from 'components/MQueryClient';
import MSnackbarProvider from 'components/MSnackbarProvider';
import { newTheme } from 'theme';

export interface Options<
  Q extends Queries = typeof queries,
  Container extends Element | DocumentFragment = HTMLElement,
> extends RenderOptions<Q, Container> {
  initialRoute: string | Partial<Location>;
}

function testRenderer<Q extends Queries = typeof queries,
  Container extends Element | DocumentFragment = HTMLElement,
>(
  ui: React.ReactElement,
  options?: Options<Q, Container>
): RenderResult<Q, Container> {
  const Wrapper = (props: React.PropsWithChildren<any>) => {
    return (
      <MemoryRouter initialEntries={ [options.initialRoute] }>
        <MQueryClient>
          <ThemeProvider theme={ newTheme }>
            <MSnackbarProvider>
              <NiceModal.Provider>
                <CssBaseline />
                {props.children}
              </NiceModal.Provider>
            </MSnackbarProvider>
          </ThemeProvider>
        </MQueryClient>
      </MemoryRouter>
    );
  };

  return render<Q, Container>(ui, { wrapper: Wrapper, ...options });
}

export default testRenderer;
