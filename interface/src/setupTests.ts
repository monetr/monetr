// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
// import '@testing-library/jest-dom';

// import { configure } from '@testing-library/react';
// import axios from 'axios';
const { beforeAll, afterEach, afterAll } = require('bun:test');

const { server } = require('@monetr/interface/testutils/server');
// import { server } from '';

const { GlobalRegistrator } = require('@happy-dom/global-registrator');
// import { GlobalRegistrator } from '@happy-dom/global-registrator';

const axios = require('axios');

const { configure } = require('@testing-library/react');

GlobalRegistrator.register();

module.exports = global.CONFIG = {
  BOOTSTRAP_CONFIG_JSON: false,
  USE_LOCAL_STORAGE: false,
  COOKIE_DOMAIN: 'app.monetr.mini',
  ENVIRONMENT: 'Testing',
  API_URL: 'https://app.monetr.mini',
  API_DOMAIN: 'app.monetr.mini',
  SENTRY_DSN: null,
};

location.href = 'http://monetr.local';

window.API = axios.create({
  baseURL: '/api',
});

configure({
  asyncUtilTimeout: 10000,
});
beforeAll(() => server.listen({
  onUnhandledRequest: 'error',
}));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
