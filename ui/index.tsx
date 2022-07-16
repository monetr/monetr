import React from 'react';
import ReactDOM from 'react-dom';
import * as Sentry from '@sentry/react';
import { Integrations } from '@sentry/tracing';

import reportWebVitals from './reportWebVitals';

import { NewClient } from 'api/api';
import axios from 'axios';
import RelayTransport from 'relay/transport';
import Root from 'Root';

import '@fontsource/roboto/300.css';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import './styles/styles.css';
import './styles/index.scss';

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

    ReactDOM.render(
      <Root />,
      document.getElementById('root')
    );
  });

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
