import '@fontsource-variable/inter';

import React from 'react';
import { createRoot } from 'react-dom/client';
import * as Sentry from '@sentry/react';
import { Integrations } from '@sentry/tracing';

import RelayTransport from '@monetr/interface/relay/transport';
import reportWebVitals from '@monetr/interface/reportWebVitals';
import Root from '@monetr/interface/Root';

import '@monetr/interface/styles/styles.css';
import '@monetr/interface/styles/index.scss';

const container = document.getElementById('root');
const root = createRoot(container);

if (window?.__MONETR__?.SENTRY_DSN) {
  Sentry.init({
    dsn: window?.__MONETR__?.SENTRY_DSN,
    transport: RelayTransport,
    integrations: [
      new Integrations.BrowserTracing({
        startTransactionOnPageLoad: false,
        startTransactionOnLocationChange: false,
        traceXHR: true,
        tracingOrigins: [
          window.location.hostname,
        ],
      }),
    ],
    release: RELEASE,
    // We recommend adjusting this value in production, or using tracesSampler
    // for finer control
    tracesSampleRate: 1.0,
    environment: window.location.hostname,
    normalizeDepth: 20,
    beforeSend(event, _) {
      // Check if it is an exception, and if so, show the report dialog
      if (event.exception) {
        try {
          Sentry.showReportDialog({ eventId: event.event_id });
        } catch (e) {
          console.error(e);
        }
      }
      return event;
    },
  });
}

root.render(
  <Root />,
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
