import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Queries, queries, render, RenderOptions, RenderResult } from '@testing-library/react';
import { SnackbarProvider, VariantType } from 'notistack';

import { newTheme } from 'theme';
import Query from 'util/query';

export interface Options<Q extends Queries = typeof queries,
  Container extends Element | DocumentFragment = HTMLElement,
  > extends RenderOptions<Q, Container> {
    initialRoute: string;
}

function testRenderer<Q extends Queries = typeof queries,
  Container extends Element | DocumentFragment = HTMLElement,
  >(
  ui: React.ReactElement,
  options?: Options<Q, Container>
): RenderResult<Q, Container> {
  const Wrapper = (props: React.PropsWithChildren<any>) => {
    const snackbarIcons: Partial<Record<VariantType, React.ReactNode>> = {
      error: <ErrorIcon className="mr-2.5" />,
      success: <DoneIcon className="mr-2.5" />,
      warning: <WarningIcon className="mr-2.5" />,
      info: <InfoIcon className="mr-2.5" />,
    };

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          queryFn: Query,
        },
      },
    });

    return (
      <MemoryRouter initialEntries={ [options.initialRoute] }>
        <QueryClientProvider client={ queryClient }>
          <ThemeProvider theme={ newTheme }>
            <LocalizationProvider dateAdapter={ AdapterMoment }>
              <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
                <NiceModal.Provider>
                  <CssBaseline />
                  { props.children }
                </NiceModal.Provider>
              </SnackbarProvider>
            </LocalizationProvider>
          </ThemeProvider>
        </QueryClientProvider>
      </MemoryRouter>
    );
  };

  return render<Q, Container>(ui, { wrapper: Wrapper, ...options });
}

export default testRenderer;
