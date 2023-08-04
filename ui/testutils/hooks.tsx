import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import NiceModal from '@ebay/nice-modal-react';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';
import { QueryClient } from '@tanstack/query-core';
import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, RenderHookResult, WrapperComponent } from '@testing-library/react-hooks';
import { SnackbarProvider, VariantType } from 'notistack';

import { newTheme } from 'theme';
import Query from 'util/query';

export interface HooksOptions {
  initialRoute: string;
}

function testRenderHook<TProps, TResult>(
  callback: (props: TProps) => TResult,
  options?: HooksOptions,
): RenderHookResult<TProps, TResult> {
  const Wrapper: WrapperComponent<TProps> = (props: React.PropsWithChildren<any>) => {
    const snackbarIcons: Partial<Record<VariantType, React.ReactNode>> = {
      error: <ErrorIcon className="mr-2.5" />,
      success: <DoneIcon className="mr-2.5" />,
      warning: <WarningIcon className="mr-2.5" />,
      info: <InfoIcon className="mr-2.5" />,
    };

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          queryFn: Query,
        },
      },
    });

    return (
      <MemoryRouter initialEntries={ [options.initialRoute] }>
        <QueryClientProvider client={ queryClient }>
          <ThemeProvider theme={ newTheme }>
            <LocalizationProvider dateAdapter={ AdapterMoment }>
              <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
                <NiceModal.Provider>
                  <CssBaseline />
                  { props.children }
                </NiceModal.Provider>
              </SnackbarProvider>
            </LocalizationProvider>
          </ThemeProvider>
        </QueryClientProvider>
      </MemoryRouter>
    );
  };

  return renderHook<TProps, TResult>(callback, {
    wrapper: Wrapper,
  });
}

export default testRenderHook;
