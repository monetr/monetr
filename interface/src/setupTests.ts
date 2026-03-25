import { expect, rs } from '@rstest/core';

import * as jestDomMatchers from '@testing-library/jest-dom/matchers';
import { cleanup, configure } from '@testing-library/react';

expect.extend(jestDomMatchers);

configure({
  asyncUtilTimeout: 10000,
});

// I dont want to see not wrapped in act stuff for stuff that is wrapped in act, pretty sure this is due to react-query
// and stuff that is happening outside of the test loop.
const _consoleError = console.error;
console.error = (...args: Parameters<typeof console.error>) => {
  if (typeof args[0] === 'string' && args[0].includes('was not wrapped in act')) {
    return;
  }
  _consoleError(...args);
};

beforeAll(() => {
  // https://github.com/jsdom/jsdom/issues/3368
  global.ResizeObserver = class ResizeObserver {
    public observe() {
      // do nothing
    }
    public unobserve() {
      // do nothing
    }
    public disconnect() {
      // do nothing
    }
  };

  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: rs.fn().mockImplementation(query => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: rs.fn(), // Deprecated
      removeListener: rs.fn(), // Deprecated
      addEventListener: rs.fn(),
      removeEventListener: rs.fn(),
      dispatchEvent: rs.fn(),
    })),
  });
});

afterEach(() => {
  cleanup();
});
