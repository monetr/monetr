// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom';

import axios from 'axios';

import { server } from './testutils/server';

module.export = global.CONFIG = {
  BOOTSTRAP_CONFIG_JSON: false,
  USE_LOCAL_STORAGE: false,
  COOKIE_DOMAIN: 'app.monetr.mini',
  ENVIRONMENT: 'Testing',
  API_URL: 'https://app.monetr.mini',
  API_DOMAIN: 'app.monetr.mini',
  SENTRY_DSN: null,
};

window.API = axios.create({
  baseURL: '/api',
});

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
