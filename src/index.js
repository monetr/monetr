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

// eslint-disable-next-line no-undef
if (CONFIG.SENTRY_DSN) {
  Sentry.init({
    // eslint-disable-next-line no-undef
    dsn: CONFIG.SENTRY_DSN,
    // eslint-disable-next-line no-undef
    release: `web-ui@${RELEASE_REVISION}`,
    integrations: [new Integrations.BrowserTracing()],
    tracesSampleRate: 1.0,
  });
}

const store = configureStore();

if (module.hot) {
  module.hot.accept()
}

ReactDOM.render(
  <React.StrictMode>
    <Provider store={ store }>
      <Sentry.ErrorBoundary fallback={"A fatal error has occurred"}>
        <Root/>
      </Sentry.ErrorBoundary>
    </Provider>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
