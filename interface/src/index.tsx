import '@fontsource-variable/inter';

import React from 'react';
import { init, reactRouterV6BrowserTracingIntegration, showReportDialog } from '@sentry/react';
import { createRoutesFromChildren, matchRoutes, useLocation, useNavigationType } from 'react-router-dom';

import Root from '@monetr/interface/Root';
import { makeSneakyFetchTransport } from '@monetr/interface/relay/transport';

import { createRoot } from 'react-dom/client';

import '@monetr/interface/styles/styles.css';
import '@monetr/interface/styles/index.scss';

const container = document.getElementById('root');
const root = createRoot(container);

if (window?.__MONETR__?.SENTRY_DSN) {
  init({
    dsn: window?.__MONETR__?.SENTRY_DSN,
    transport: makeSneakyFetchTransport,
    integrations: [
      reactRouterV6BrowserTracingIntegration({
        useEffect: React.useEffect,
        useLocation,
        useNavigationType,
        createRoutesFromChildren,
        matchRoutes,
      }),
    ],
    release: RELEASE,
    // We recommend adjusting this value in production, or using tracesSampler
    // for finer control
    tracesSampleRate: 1.0,
    sampleRate: 1.0,
    environment: window.location.hostname,
    normalizeDepth: 20,
    beforeSend(event, _) {
      // If the exception's stack trace includes a single line that is from any kind of browser extension then we don't
      // want to hear about the error in sentry. It isn't helpful and creates noise.
      if ((event?.exception?.values ?? []).find(exception =>
        (exception?.stacktrace?.frames ?? []).find(stacktrace => (stacktrace?.filename ?? '').includes('extension://')),
      )) {
        return null;
      }

      // Check if it is an exception, and if so, show the report dialog
      if (event.exception) {
        try {
          showReportDialog({ eventId: event.event_id });
        } catch (e) {
          console.error(e);
        }
      }
      return event;
    },
  });
}

root.render(<Root />);
