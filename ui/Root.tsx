import { LocalizationProvider } from '@mui/lab';
import AdapterMoment from '@mui/lab/AdapterMoment';
import { createTheme, CssBaseline, ThemeProvider } from '@mui/material';
import Application from 'Application';
import GlobalFooter from 'components/GlobalFooter';
import React from 'react';
import { Provider } from 'react-redux';
import { BrowserRouter as Router } from 'react-router-dom';
import { store } from 'store';
import { IconVariant, SnackbarProvider } from 'notistack';
import * as Sentry from '@sentry/react';
import ErrorIcon from '@mui/icons-material/Error';
import DoneIcon from '@mui/icons-material/Done';
import WarningIcon from '@mui/icons-material/Warning';
import InfoIcon from '@mui/icons-material/Info';

export default function Root(): JSX.Element {
  const darkMode = window.localStorage.getItem('darkMode') === 'true';
  const theme = createTheme({
    shape: {
      borderRadius: 10,
    },
    palette: {
      mode: darkMode ? 'dark' : 'light',
      primary: {
        main: darkMode ? '#712ddd' : '#4E1AA0',
        contrastText: '#FFFFFF',
      },
      secondary: {
        main: '#FF5798',
        contrastText: '#FFFFFF',
      },
      background: {
        default: darkMode ? '#2f2f2f' : '#FFFFFF',
      }
    }
  });

  const snackbarIcons: Partial<IconVariant> = {
    error: <ErrorIcon className="mr-2.5"/>,
    success: <DoneIcon className="mr-2.5"/>,
    warning: <WarningIcon className="mr-2.5"/>,
    info: <InfoIcon className="mr-2.5"/>,
  };

  return (
    <React.StrictMode>
      <Sentry.ErrorBoundary>
        <Provider store={ store }>
          <Router>
            <ThemeProvider theme={ theme }>
              <LocalizationProvider dateAdapter={ AdapterMoment }>
                <SnackbarProvider maxSnack={ 5 } iconVariant={ snackbarIcons }>
                  <CssBaseline/>
                  <Application/>
                  <GlobalFooter/>
                </SnackbarProvider>
              </LocalizationProvider>
            </ThemeProvider>
          </Router>
        </Provider>
      </Sentry.ErrorBoundary>
    </React.StrictMode>
  )
}