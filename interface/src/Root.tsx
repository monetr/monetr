import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';

import Monetr from './monetr';
import { newTheme } from './theme';

import MQueryClient from './components/MQueryClient';
import MSnackbarProvider from './components/MSnackbarProvider';

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
