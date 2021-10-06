import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux'
import reportWebVitals from './reportWebVitals';
import Root from "./root";
import configureStore from './store';
import "./styles/styles.css";
import './styles/index.scss';
import * as Sentry from "@sentry/react";
import { Integrations } from "@sentry/tracing";
import { BrowserRouter as Router } from "react-router-dom";
import { createMuiTheme, MuiThemeProvider, Typography } from "@material-ui/core";

// eslint-disable-next-line no-undef
if (CONFIG.SENTRY_DSN) {
  Sentry.init({
    // eslint-disable-next-line no-undef
    dsn: CONFIG.SENTRY_DSN,
    // eslint-disable-next-line no-undef
    release: `web-ui@${ RELEASE_REVISION }`,
    integrations: [
      new Integrations.BrowserTracing({
        tracingOrigins: [
          // eslint-disable-next-line no-undef
          CONFIG.API_DOMAIN,
        ]
      }),
    ],
    tracesSampleRate: 1,
    autoSessionTracking: true
  });
}

const store = configureStore();

if (module.hot) {
  module.hot.accept()
}

const theme = createMuiTheme({
  palette: {
    primary: {
      main: '#4E1AA0',
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: '#FF5798',
      contrastText: '#FFFFFF',
    }
  }
});

ReactDOM.render(
  <React.StrictMode>
    <Sentry.ErrorBoundary fallback={ "A fatal error has occurred" }>
      <Provider store={ store }>
        <Router>
          <MuiThemeProvider theme={ theme }>
            <Root/>
            <Typography
              className="absolute bottom-1 w-full text-center opacity-30 inline"
            >
              {/* eslint-disable-next-line no-undef */ }
              © { new Date().getFullYear() } monetr LLC - { RELEASE_REVISION.slice(0, 8) }
            </Typography>
          </MuiThemeProvider>
        </Router>
      </Provider>
    </Sentry.ErrorBoundary>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
