import NiceModal from '@ebay/nice-modal-react';
import { ErrorBoundary } from '@sentry/react';
import { BrowserRouter } from 'react-router-dom';

import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';
import PullToRefresh from '@monetr/interface/components/PullToRefresh';
import { TooltipProvider } from '@monetr/interface/components/Tooltip';
import Monetr from '@monetr/interface/monetr';

export default function Root(): JSX.Element {
  return (
    <BrowserRouter>
      <ErrorBoundary>
        <MQueryClient>
          <MSnackbarProvider>
            <TooltipProvider>
              <NiceModal.Provider>
                <PullToRefresh />
                <Monetr />
              </NiceModal.Provider>
            </TooltipProvider>
          </MSnackbarProvider>
        </MQueryClient>
      </ErrorBoundary>
    </BrowserRouter>
  );
}
