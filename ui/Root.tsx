import React from 'react';
import {
  QueryClient,
  QueryClientProvider, QueryFunctionContext, QueryKey,
} from 'react-query';
import { Provider } from 'react-redux';
import { BrowserRouter as Router } from 'react-router-dom';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { LocalizationProvider } from '@mui/lab';
import AdapterMoment from '@mui/lab/AdapterMoment';
import { CssBaseline, ThemeProvider } from '@mui/material';
import * as Sentry from '@sentry/react';
import { IconVariant, SnackbarProvider } from 'notistack';
import axios from 'axios';

import Application from 'Application';
import GlobalFooter from 'components/GlobalFooter';
import { store } from 'store';
import theme from 'theme';

export default function Root(): JSX.Element {
  const snackbarIcons: Partial<IconVariant> = {
    error: <ErrorIcon className="mr-2.5" />,
    success: <DoneIcon className="mr-2.5" />,
    warning: <WarningIcon className="mr-2.5" />,
    info: <InfoIcon className="mr-2.5" />,
  };

  async function queryFn<T = unknown, TQueryKey extends QueryKey = QueryKey>(
    context: QueryFunctionContext<TQueryKey>,
  ): Promise<T> {
    const { data } = await axios.get<T>(`/api${ context.queryKey[0] }`)
      .catch(result => {
        switch (result.response.status) {
          case 500:
            throw result;
          default:
            return result.response;
        }
      });
    return data;
  }

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 5 * 60 * 1000, // 5 minute default stale time,
        queryFn: queryFn,
      },
    },
  });

  return (
    <React.StrictMode>
      <Sentry.ErrorBoundary>
        <QueryClientProvider client={ queryClient }>
          <Provider store={ store }>
            <Router>
              <ThemeProvider theme={ theme }>
                <LocalizationProvider dateAdapter={ AdapterMoment }>
                  <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
                    <CssBaseline />
                    <Application />
                    <GlobalFooter />
                  </SnackbarProvider>
                </LocalizationProvider>
              </ThemeProvider>
            </Router>
          </Provider>
        </QueryClientProvider>
      </Sentry.ErrorBoundary>
    </React.StrictMode>
  );
}
