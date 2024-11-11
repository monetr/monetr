import '@fontsource-variable/inter';

import React, { useEffect } from 'react';
import NiceModal from '@ebay/nice-modal-react';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { INITIAL_VIEWPORTS } from '@storybook/addon-viewport';
import type { Preview } from '@storybook/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider, VariantType } from 'notistack';

import { TooltipProvider } from '@monetr/interface/components/Tooltip';
import { newTheme } from '@monetr/interface/theme';
import Query from '@monetr/interface/util/query';

import { withRouter } from 'storybook-addon-react-router-v6';

import '@monetr/interface/styles/styles.css';
import '@monetr/interface/styles/index.scss';
import './preview.css';

export const useTheme = (StoryFn: () => unknown) => {
  useEffect(() => {
    document.querySelector('html')?.setAttribute('class', 'dark');
  }, []);

  return StoryFn();
};

const preview: Preview = {
  decorators: [
    useTheme,
    (Story, _context) => {
      const snackbarIcons: Partial<Record<VariantType, React.ReactNode>> = {
        error: <ErrorIcon className='mr-2.5' />,
        success: <DoneIcon className='mr-2.5' />,
        warning: <WarningIcon className='mr-2.5' />,
        info: <InfoIcon className='mr-2.5' />,
      };

      const queryClient = new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 10 * 60 * 1000, // 10 minute default stale time,
            queryFn: Query,
          },
        },
      });

      return (
        <QueryClientProvider client={ queryClient }>
          <ThemeProvider theme={ newTheme }>
            <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
              <TooltipProvider>
                <NiceModal.Provider>
                  <CssBaseline />
                  <Story />
                </NiceModal.Provider>
              </TooltipProvider>
            </SnackbarProvider>
          </ThemeProvider>
        </QueryClientProvider>
      );
    },
    withRouter,
  ],
  args: {

  },
  parameters: {
    // screenshot: {
    //   viewport: {
    //     width: 1280,
    //     height: 720,
    //     isMobile: false,
    //     hasTouch: false,
    //   },
    //   delay: 3000,
    // } as ScreenshotOptions,
    viewport: {
      viewports: {
        desktop: {
          name: 'Desktop',
          styles: {
            width: '1280px',
            height: '720px',
          },
        },
        ...INITIAL_VIEWPORTS,
      },
    },
    actions: { argTypesRegex: '^on[A-Z].*' },
    // controls: {
    //   matchers: {
    //     color: /(background|color)$/i,
    //     date: /Date$/,
    //   },
    // },
  },
};

export default preview;

// // TODO This will sometimes crash chrome for some reason?
// export const globalTypes = {
//   theme: {
//     name: 'Toggle theme',
//     description: 'Global theme for components',
//     defaultValue: 'dark',
//     toolbar: {
//       icon: 'circlehollow',
//       items: ['dark'], // 'light',
//       showName: true,
//       dynamicTitle: true,
//     },
//   },
// };
