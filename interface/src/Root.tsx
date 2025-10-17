import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import * as Sentry from '@sentry/react';

import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';
import PullToRefresh from '@monetr/interface/components/PullToRefresh';
import { TooltipProvider } from '@monetr/interface/components/Tooltip';
import Monetr from '@monetr/interface/monetr';
import { newTheme } from '@monetr/interface/theme';

export default function Root(): JSX.Element {
  return (
    <BrowserRouter>
      <Sentry.ErrorBoundary>
        <MQueryClient>
          <ThemeProvider theme={newTheme}>
            <MSnackbarProvider>
              <TooltipProvider>
                <NiceModal.Provider>
                  <CssBaseline />
                  <PullToRefresh />
                  <Monetr />
                </NiceModal.Provider>
              </TooltipProvider>
            </MSnackbarProvider>
          </ThemeProvider>
        </MQueryClient>
      </Sentry.ErrorBoundary>
    </BrowserRouter>
  );
}
