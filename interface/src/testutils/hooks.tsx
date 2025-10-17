import type React from 'react';
import { MemoryRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { renderHook, type RenderHookResult, type WrapperComponent } from '@testing-library/react-hooks';

import MQueryClient from '@monetr/interface/components/MQueryClient';
import MSnackbarProvider from '@monetr/interface/components/MSnackbarProvider';

import { newTheme } from 'theme';

export interface HooksOptions<TProps> {
  initialRoute: string;
  initialProps?: TProps;
}

function testRenderHook<TProps, TResult>(
  callback: (props: TProps) => TResult,
  options?: HooksOptions<TProps>,
): RenderHookResult<TProps, TResult> {
  const Wrapper: WrapperComponent<TProps> = (props: React.PropsWithChildren<any>) => {
    return (
      <MemoryRouter
        initialEntries={[options.initialRoute]}
        future={{ v7_startTransition: false, v7_relativeSplatPath: false }}
      >
        <MQueryClient>
          <ThemeProvider theme={newTheme}>
            <MSnackbarProvider>
              <NiceModal.Provider>
                <CssBaseline />
                {props.children}
              </NiceModal.Provider>
            </MSnackbarProvider>
          </ThemeProvider>
        </MQueryClient>
      </MemoryRouter>
    );
  };

  return renderHook<TProps, TResult>(callback, {
    wrapper: Wrapper,
    initialProps: options?.initialProps,
  });
}

export default testRenderHook;
