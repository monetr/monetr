import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';

import MQueryClient from 'components/MQueryClient';
import MSnackbarProvider from 'components/MSnackbarProvider';
import Monetr from 'monetr';
import { newTheme } from 'theme';

export default function Root(): JSX.Element {
  // <Sentry.ErrorBoundary>
  return (
    <BrowserRouter>
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
    </BrowserRouter>
  );
}
