import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { LocalizationProvider } from '@mui/lab';
import AdapterMoment from '@mui/lab/AdapterMoment';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { queries, Queries, render, RenderOptions, RenderResult } from '@testing-library/react';
import Application from 'Application';
import GlobalFooter from 'components/GlobalFooter';
import { IconVariant, SnackbarProvider } from 'notistack';
import React from 'react';
import { Provider } from 'react-redux';
import { AppStore, configureStore, store } from 'store';
import theme from 'theme';

export interface Options<Q extends Queries = typeof queries,
  Container extends Element | DocumentFragment = HTMLElement,
  > extends RenderOptions<Q, Container> {
  store?: AppStore;
}

function testRenderer<Q extends Queries = typeof queries,
  Container extends Element | DocumentFragment = HTMLElement,
  >(
  ui: React.ReactElement,
  options?: Options<Q, Container>
): RenderResult<Q, Container> {
  const { store: {}, ...testLibraryOptions } = {
    store: {},
    ...options
  };
  const Wrapper = (props: React.PropsWithChildren<any>) => {
    const store = options?.store || configureStore();

    const snackbarIcons: Partial<IconVariant> = {
      error: <ErrorIcon className="mr-2.5"/>,
      success: <DoneIcon className="mr-2.5"/>,
      warning: <WarningIcon className="mr-2.5"/>,
      info: <InfoIcon className="mr-2.5"/>,
    };

    return (
      <Provider store={ store }>
        <ThemeProvider theme={ theme }>
          <LocalizationProvider dateAdapter={ AdapterMoment }>
            <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
              <CssBaseline/>
              { props.children }
            </SnackbarProvider>
          </LocalizationProvider>
        </ThemeProvider>
      </Provider>
    );
  }

  return render<Q, Container>(ui, { wrapper: Wrapper, ...testLibraryOptions })
}

export default testRenderer;
