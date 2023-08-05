import React from 'react';
import { BrowserRouter as Router } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';
import * as Sentry from '@sentry/react';

import MQueryClient from 'components/MQueryClient';
import MSnackbarProvider from 'components/MSnackbarProvider';
import Monetr from 'monetr';
import theme, { newTheme } from 'theme';

export default function Root(): JSX.Element {

  // <Sentry.ErrorBoundary>
  return (
    <React.StrictMode>
      <Router>
        <MQueryClient>
          <ThemeProvider theme={ newTheme }>
            <LocalizationProvider dateAdapter={ AdapterMoment }>
              <MSnackbarProvider>
                <NiceModal.Provider>
                  <CssBaseline />
                  <Monetr />
                </NiceModal.Provider>
              </MSnackbarProvider>
            </LocalizationProvider>
          </ThemeProvider>
        </MQueryClient>
      </Router>
    </React.StrictMode>
  );
}
