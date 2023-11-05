import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';

import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';
import Monetr from '@monetr/interface/monetr';
import { newTheme } from '@monetr/interface/theme';

export default function Root(): JSX.Element {
  // <Sentry.ErrorBoundary>
  return (
    <BrowserRouter>
      <MQueryClient>
        <ThemeProvider theme={ newTheme }>
          <MSnackbarProvider>
            <NiceModal.Provider>
              <CssBaseline />
              <Monetr />
            </NiceModal.Provider>
          </MSnackbarProvider>
        </ThemeProvider>
      </MQueryClient>
    </BrowserRouter>
  );
}
