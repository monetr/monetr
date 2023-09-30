import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';
import { renderHook, RenderHookResult, WrapperComponent } from '@testing-library/react-hooks';

import MQueryClient from 'components/MQueryClient';
import MSnackbarProvider from 'components/MSnackbarProvider';
import { newTheme } from 'theme';

export interface HooksOptions {
  initialRoute: string;
}

function testRenderHook<TProps, TResult>(
  callback: (props: TProps) => TResult,
  options?: HooksOptions,
): RenderHookResult<TProps, TResult> {
  const Wrapper: WrapperComponent<TProps> = (props: React.PropsWithChildren<any>) => {
    return (
      <MemoryRouter initialEntries={ [options.initialRoute] }>
        <MQueryClient>
          <ThemeProvider theme={ newTheme }>
            <LocalizationProvider dateAdapter={ AdapterMoment }>
              <MSnackbarProvider>
                <NiceModal.Provider>
                  <CssBaseline />
                  { props.children }
                </NiceModal.Provider>
              </MSnackbarProvider>
            </LocalizationProvider>
          </ThemeProvider>
        </MQueryClient>
      </MemoryRouter>
    );
  };

  return renderHook<TProps, TResult>(callback, {
    wrapper: Wrapper,
  });
}

export default testRenderHook;