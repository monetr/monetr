import { expect } from '@rstest/core';
import * as jestDomMatchers from '@testing-library/jest-dom/matchers';
import { cleanup, configure } from '@testing-library/react';

expect.extend(jestDomMatchers);

configure({
  asyncUtilTimeout: 10000,
});

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
    value: rstest.fn().mockImplementation(query => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: rstest.fn(), // Deprecated
      removeListener: rstest.fn(), // Deprecated
      addEventListener: rstest.fn(),
      removeEventListener: rstest.fn(),
      dispatchEvent: rstest.fn(),
    })),
  });
});

afterEach(() => {
  cleanup();
});
