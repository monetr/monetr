import '@fontsource-variable/inter';

import React from 'react';
import { QueryClient, QueryClientProvider, QueryFunctionContext, QueryKey } from 'react-query';
import NiceModal from '@ebay/nice-modal-react';
import DoneIcon from '@mui/icons-material/Done';
import ErrorIcon from '@mui/icons-material/Error';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';
import { INITIAL_VIEWPORTS } from '@storybook/addon-viewport';
import { useEffect, useGlobals } from '@storybook/addons';
import type { Preview } from '@storybook/react';
import axios from 'axios';
import { SnackbarProvider, VariantType } from 'notistack';

import theme, { newTheme } from '../ui/theme';

import { initialize, mswLoader } from 'msw-storybook-addon';
import { withRouter } from 'storybook-addon-react-router-v6';

import '../ui/styles/styles.css';
import './preview.css';

initialize({
  onUnhandledRequest: 'bypass',
});

window.API = axios.create({
  baseURL: '/api',
});

export const useTheme = (StoryFn: () => unknown) => {
  const [{ theme }] = useGlobals();

  useEffect(() => {
    document.querySelector('html')?.setAttribute('class', theme || 'dark');
  }, [theme]);

  return StoryFn();
};

const preview: Preview = {
  decorators: [
    useTheme,
    (Story, _context) => {
      const snackbarIcons: Partial<Record<VariantType, React.ReactNode>> = {
        error: <ErrorIcon className="mr-2.5" />,
        success: <DoneIcon className="mr-2.5" />,
        warning: <WarningIcon className="mr-2.5" />,
        info: <InfoIcon className="mr-2.5" />,
      };

      async function queryFn<T = unknown, TQueryKey extends QueryKey = QueryKey>(
        context: QueryFunctionContext<TQueryKey>,
      ): Promise<T> {
        const { data } = await axios.request<T>({
          url: `/api${context.queryKey[0]}`,
          method: context.queryKey.length === 1 ? 'GET' : 'POST',
          params: context.pageParam && {
            offset: context.pageParam,
          },
          data: context.queryKey.length === 2 && context.queryKey[1],
        })
          .catch(result => {
            switch (result.response.status) {
              case 500: // Internal Server Error
                throw result;
              default:
                return result.response;
            }
          });
        return data;
      }

      const queryClient = new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 10 * 60 * 1000, // 10 minute default stale time,
            queryFn: queryFn,
          },
        },
      });

      return (
        <QueryClientProvider client={ queryClient }>
          <ThemeProvider theme={ newTheme }>
            <LocalizationProvider dateAdapter={ AdapterMoment }>
              <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
                <NiceModal.Provider>
                  <CssBaseline />
                  <Story />
                </NiceModal.Provider>
              </SnackbarProvider>
            </LocalizationProvider>
          </ThemeProvider>
        </QueryClientProvider>
      );
    },
    withRouter,
  ],
  args: {

  },
  parameters: {
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
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/,
      },
    },
  },
  loaders: [mswLoader],
};

export default preview;

// TODO This will sometimes crash chrome for some reason?
export const globalTypes = {
  theme: {
    name: 'Toggle theme',
    description: 'Global theme for components',
    defaultValue: 'dark',
    toolbar: {
      icon: 'circlehollow',
      items: ['dark'], // 'light',
      showName: true,
      dynamicTitle: true,
    },
  },
};
