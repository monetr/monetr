import { LocalizationProvider } from '@mui/lab';
import AdapterMoment from '@mui/lab/AdapterMoment';
import { createTheme, CssBaseline, ThemeProvider } from '@mui/material';
import Application from 'Application';
import GlobalFooter from 'components/GlobalFooter';
import React from 'react';
import { Provider } from 'react-redux';
import { BrowserRouter as Router } from 'react-router-dom';
import { store } from 'store';
import * as Sentry from '@sentry/react';

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

  return (
    <React.StrictMode>
      <Sentry.ErrorBoundary>
        <Provider store={ store }>
          <Router>
            <ThemeProvider theme={ theme }>
              <LocalizationProvider dateAdapter={ AdapterMoment }>
                <CssBaseline/>
                <Application/>
                <GlobalFooter/>
              </LocalizationProvider>
            </ThemeProvider>
          </Router>
        </Provider>
      </Sentry.ErrorBoundary>
    </React.StrictMode>
  )
}