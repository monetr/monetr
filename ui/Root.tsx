import clsx from 'clsx';
import React from 'react';
import {
  QueryClientProvider,
} from 'react-query';
import { BrowserRouter as Router } from 'react-router-dom';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { LocalizationProvider } from '@mui/x-date-pickers'
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';
import { CssBaseline, ThemeProvider } from '@mui/material';
import * as Sentry from '@sentry/react';
import { IconVariant, SnackbarProvider } from 'notistack';
import NiceModal from '@ebay/nice-modal-react';

import Application from 'Application';
import theme from 'theme';
import createQueryClient from 'util/createQueryClient';

export default function Root(): JSX.Element {
  const snackbarIcons: Partial<IconVariant> = {
    error: <ErrorIcon className="mr-2.5" />,
    success: <DoneIcon className="mr-2.5" />,
    warning: <WarningIcon className="mr-2.5" />,
    info: <InfoIcon className="mr-2.5" />,
  };

  const queryClient = createQueryClient();

  return (
    <div className={ clsx('w-full h-full bg-purple-900', {
      'dark': theme.palette.mode === 'dark',
    })}>
      <React.StrictMode>
        <Sentry.ErrorBoundary>
          <QueryClientProvider client={ queryClient }>
            <Router>
              <ThemeProvider theme={ theme }>
                <LocalizationProvider dateAdapter={ AdapterMoment }>
                  <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
                    <NiceModal.Provider>
                      <CssBaseline />
                      <Application />
                    </NiceModal.Provider>
                  </SnackbarProvider>
                </LocalizationProvider>
              </ThemeProvider>
            </Router>
          </QueryClientProvider>
        </Sentry.ErrorBoundary>
      </React.StrictMode>
    </div>
  );
}
