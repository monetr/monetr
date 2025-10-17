import type React from 'react';
import { type Location, MemoryRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { type Queries, type queries, render, type RenderOptions, type RenderResult } from '@testing-library/react';
import type { AxiosInstance } from 'axios';

import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';
import { TooltipProvider } from '@monetr/interface/components/Tooltip';
import { newTheme } from '@monetr/interface/theme';

export interface Options<Q extends Queries = typeof queries, Container extends Element | DocumentFragment = HTMLElement>
  extends RenderOptions<Q, Container> {
  initialRoute: string | Partial<Location>;
  client?: AxiosInstance;
}

function testRenderer<Q extends Queries = typeof queries, Container extends Element | DocumentFragment = HTMLElement>(
  ui: React.ReactElement,
  options?: Options<Q, Container>,
): RenderResult<Q, Container> {
  const Wrapper = (props: React.PropsWithChildren<any>) => {
    return (
      <MemoryRouter initialEntries={[options.initialRoute]}>
        <MQueryClient client={options.client}>
          <ThemeProvider theme={newTheme}>
            <MSnackbarProvider>
              <TooltipProvider>
                <NiceModal.Provider>
                  <CssBaseline />
                  {props.children}
                </NiceModal.Provider>
              </TooltipProvider>
            </MSnackbarProvider>
          </ThemeProvider>
        </MQueryClient>
      </MemoryRouter>
    );
  };

  return render<Q, Container>(ui, { wrapper: Wrapper, ...options });
}

export default testRenderer;
