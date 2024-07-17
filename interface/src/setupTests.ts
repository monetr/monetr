// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom

import * as matchers from '@testing-library/jest-dom/matchers';
import { cleanup, configure  } from '@testing-library/react';

// import { server } from '@monetr/interface/testutils/server';
import { GlobalRegistrator } from '@happy-dom/global-registrator';
import { afterEach, expect } from 'bun:test';

GlobalRegistrator.register();;

expect.extend(matchers as any);

// import { Window } from 'happy-dom';
//
// const window = new Window();
// const document = window.document;
// global.document = document;
// global.window = window;

// module.export = global.CONFIG = {
//   BOOTSTRAP_CONFIG_JSON: false,
//   USE_LOCAL_STORAGE: false,
//   COOKIE_DOMAIN: 'app.monetr.mini',
//   ENVIRONMENT: 'Testing',
//   API_URL: 'https://app.monetr.mini',
//   API_DOMAIN: 'app.monetr.mini',
//   SENTRY_DSN: null,
// };

// axios.defaults.adapter = 'http';
// location.href = 'https://monetr.local';

configure({
  asyncUtilTimeout: 10000,
});
// beforeAll(() => server.listen({
//   onUnhandledRequest: 'error',
// }));
afterEach(() => {
  cleanup();
});
// afterAll(() => server.close());
