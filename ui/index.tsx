import { Integrations } from '@sentry/tracing';
import axios from 'axios';
import React from 'react';
import ReactDOM from 'react-dom';
import RelayTransport from 'relay/transport';
import Root from 'Root';
import reportWebVitals from './reportWebVitals';
import './styles/styles.css';
import './styles/index.scss';
import * as Sentry from '@sentry/react';

axios.get('/api/sentry')
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
            ]
          })
        ],
        release: RELEASE,
        // We recommend adjusting this value in production, or using tracesSampler
        // for finer control
        tracesSampleRate: 1.0,
        environment: window.location.hostname,
        normalizeDepth: 20,
        beforeSend(event, hint) {
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
    ReactDOM.render(
      <Root/>,
      document.getElementById('root')
    );
  });

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
