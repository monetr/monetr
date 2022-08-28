import React from 'react';
import { unstable_HistoryRouter as HistoryRouter } from 'react-router-dom';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { LocalizationProvider } from '@mui/lab';
import AdapterMoment from '@mui/lab/AdapterMoment';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { Queries, queries, render, RenderOptions, RenderResult } from '@testing-library/react';
import { IconVariant, SnackbarProvider } from 'notistack';
import { createMemoryHistory, MemoryHistory } from 'history';

import theme from 'theme';

export interface Options<Q extends Queries = typeof queries,
  Container extends Element | DocumentFragment = HTMLElement,
  > extends RenderOptions<Q, Container> {
  history?: MemoryHistory;
}

function testRenderer<Q extends Queries = typeof queries,
  Container extends Element | DocumentFragment = HTMLElement,
  >(
  ui: React.ReactElement,
  options?: Options<Q, Container>
): RenderResult<Q, Container> {
  const Wrapper = (props: React.PropsWithChildren<any>) => {
    const history = options?.history || createMemoryHistory();

    const snackbarIcons: Partial<IconVariant> = {
      error: <ErrorIcon className="mr-2.5" />,
      success: <DoneIcon className="mr-2.5" />,
      warning: <WarningIcon className="mr-2.5" />,
      info: <InfoIcon className="mr-2.5" />,
    };

    return (
      <HistoryRouter history={ history }>
        <ThemeProvider theme={ theme }>
          <LocalizationProvider dateAdapter={ AdapterMoment }>
            <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
              <CssBaseline />
              { props.children }
            </SnackbarProvider>
          </LocalizationProvider>
        </ThemeProvider>
      </HistoryRouter>
    );
  };

  return render<Q, Container>(ui, { wrapper: Wrapper, ...options });
}

export default testRenderer;
