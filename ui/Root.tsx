import axios from 'axios';
import React from 'react';
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
import {
  useQuery,
  useMutation,
  useQueryClient,
  QueryClient,
  QueryClientProvider,
} from 'react-query';

import Application from 'Application';
import GlobalFooter from 'components/GlobalFooter';
import { IconVariant, SnackbarProvider } from 'notistack';
import { store } from 'store';
import theme from 'theme';

export default function Root(): JSX.Element {
  const snackbarIcons: Partial<IconVariant> = {
    error: <ErrorIcon className="mr-2.5" />,
    success: <DoneIcon className="mr-2.5" />,
    warning: <WarningIcon className="mr-2.5" />,
    info: <InfoIcon className="mr-2.5" />,
  };

  const defaultQueryFn = async ({ queryKey }) => {
    const { data } = await axios.get(`${queryKey[0]}`);
    return data;
  };


  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        queryFn: defaultQueryFn,
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
