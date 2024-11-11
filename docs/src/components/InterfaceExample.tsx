/* eslint-disable max-len */
'use client';

import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import MockAdapter from 'axios-mock-adapter';

import monetrClient from '@monetr/interface/api/api';
import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';
import { TooltipProvider } from '@monetr/interface/components/Tooltip';
import Monetr from '@monetr/interface/monetr';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import { newTheme } from '@monetr/interface/theme';

export default function InterfaceExample(): JSX.Element {
  const initialRoute = '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions';
  const mockAxios = new MockAdapter(monetrClient);
  apiSampleResponses(mockAxios);

  return (
    <div className='w-full h-full rounded-2xl mt-8 shadow-2xl z-10 backdrop-blur-md bg-black/90 pointer-events-none select-none max-w-[1280px] max-h-[720px] aspect-video-vertical md:aspect-video'>
      <MemoryRouter initialEntries={ [initialRoute] }>
        <MQueryClient>
          <ThemeProvider theme={ newTheme }>
            <MSnackbarProvider>
              <TooltipProvider>
                <NiceModal.Provider>
                  <CssBaseline />
                  <Monetr />
                </NiceModal.Provider>
              </TooltipProvider>
            </MSnackbarProvider>
          </ThemeProvider>
        </MQueryClient>
      </MemoryRouter>
    </div>
  );
}
