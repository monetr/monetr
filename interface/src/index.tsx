import '@fontsource-variable/inter';

import React from 'react';
import { createRoot } from 'react-dom/client';
import * as Sentry from '@sentry/react';
import { Integrations } from '@sentry/tracing';
import axios from 'axios';

import reportWebVitals from './reportWebVitals';

import { NewClient } from 'api/api';
import RelayTransport from 'relay/transport';
import Root from 'Root';

import './styles/styles.css';
import './styles/index.scss';

const container = document.getElementById('root');
const root = createRoot(container);

axios.get('/api/sentry', {
  // When the UI initially loads, it tries to talk to the API to see if sentry should be setup. If it should then it
  // tries to do that here. But if it cannot then nothing happens. We want to have a 1 second timeout for this request
  // to not hurt user experience if things are running slowly.
  timeout: 1000,
})
  .then(result => {
    if (result.data.dsn) {
      Sentry.init({
        dsn: result.data.dsn,
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
            Sentry.showReportDialog({ eventId: event.event_id });
          }
          return event;
        },
      });
    }
  })
  .finally(() => {
    window.API = NewClient({
      baseURL: '/api',
    });

    root.render(
      <Root />,
    );
  });

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
