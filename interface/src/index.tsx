import '@fontsource-variable/inter';

import React from 'react';
import { createRoot } from 'react-dom/client';
import { createRoutesFromChildren, matchRoutes, useLocation, useNavigationType } from 'react-router-dom';
import * as Sentry from '@sentry/react';

import { makeSneakyFetchTransport } from '@monetr/interface/relay/transport';
import Root from '@monetr/interface/Root';

import '@monetr/interface/styles/styles.css';
import '@monetr/interface/styles/index.scss';

const container = document.getElementById('root');
const root = createRoot(container);

if (window?.__MONETR__?.SENTRY_DSN) {
  Sentry.init({
    dsn: window?.__MONETR__?.SENTRY_DSN,
    transport: makeSneakyFetchTransport,
    integrations: [
      Sentry.reactRouterV6BrowserTracingIntegration({
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
