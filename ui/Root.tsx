import React from 'react';
import {
  QueryClient,
  QueryClientProvider, QueryFunctionContext, QueryKey,
} from 'react-query';
import { BrowserRouter as Router } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';
import * as Sentry from '@sentry/react';
import axios from 'axios';
import { SnackbarProvider, VariantType } from 'notistack';

import Monetr from 'monetr';
import theme, { newTheme } from 'theme';

export default function Root(): JSX.Element {
  const snackbarIcons: Partial<Record<VariantType, React.ReactNode>> = {
    error: <ErrorIcon className="mr-2.5" />,
    success: <DoneIcon className="mr-2.5" />,
    warning: <WarningIcon className="mr-2.5" />,
    info: <InfoIcon className="mr-2.5" />,
  };

  async function queryFn<T = unknown, TQueryKey extends QueryKey = QueryKey>(
    context: QueryFunctionContext<TQueryKey>,
  ): Promise<T> {
    const { data } = await axios.request<T>({
      url: `/api${ context.queryKey[0] }`,
      method: context.queryKey.length === 1 ? 'GET' : 'POST',
      params: context.pageParam && {
        offset: context.pageParam,
      },
      data: context.queryKey.length === 2 && context.queryKey[1],
    })
      .catch(result => {
        switch (result.response.status) {
          case 404:
          case 500: // Internal Server Error
          case 502:
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
        staleTime: 10 * 60 * 1000, // 10 minute default stale time,
        queryFn: queryFn,
      },
    },
  });

  // <Sentry.ErrorBoundary>
  return (
    <React.StrictMode>
      <Router>
        <QueryClientProvider client={ queryClient }>
          <ThemeProvider theme={ newTheme }>
            <LocalizationProvider dateAdapter={ AdapterMoment }>
              <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
                <NiceModal.Provider>
                  <CssBaseline />
                  <Monetr />
                </NiceModal.Provider>
              </SnackbarProvider>
            </LocalizationProvider>
          </ThemeProvider>
        </QueryClientProvider>
      </Router>
    </React.StrictMode>
  );
}
