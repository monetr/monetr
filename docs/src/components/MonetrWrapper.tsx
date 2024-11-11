
import React, { ReactElement } from 'react';
import { MemoryRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';
import { TooltipProvider } from '@monetr/interface/components/Tooltip';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import { newTheme } from '@monetr/interface/theme';

export interface MonetrWrapperProps {
  initialRoute?: string;
  children: ReactElement;
}

export default function MonetrWrapper(props: MonetrWrapperProps): JSX.Element {
  const mockAxios = new MockAdapter(monetrClient);
  apiSampleResponses(mockAxios);
  return (
    <MemoryRouter initialEntries={ [props?.initialRoute || '/'] }>
      <MQueryClient>
        <ThemeProvider theme={ newTheme }>
          <MSnackbarProvider>
            <TooltipProvider>
              <NiceModal.Provider>
                <CssBaseline />
                { props.children }
              </NiceModal.Provider>
            </TooltipProvider>
          </MSnackbarProvider>
        </ThemeProvider>
      </MQueryClient>
    </MemoryRouter>
  );
}
