import path from 'node:path';

import { defineConfig } from '@rstest/core';

const interfaceSource = path.resolve(import.meta.dirname, 'src');

export default defineConfig({
  globals: true,
  testEnvironment: 'jsdom',
  setupFiles: ['./src/setupTests.ts'],
  include: ['src/**/*.spec.{ts,tsx}'],
  exclude: ['node_modules'],
  source: {
    define: {
      // React 18's act() checks this global to determine if it's in a test
      // environment. Setting it here via define ensures it's baked into the
      // compiled bundle, regardless of module context separation.
      'globalThis.IS_REACT_ACT_ENVIRONMENT': 'true',
    },
  },
  resolve: {
    alias: {
      '@monetr/interface': interfaceSource,
      // @testing-library/react-hooks auto-detects the renderer via a dynamic
      // require, which rspack cannot statically analyze. Since react-test-renderer
      // is present, it picks the native renderer and fails. Force the DOM renderer.
      '@testing-library/react-hooks': '@testing-library/react-hooks/dom',
    },
  },
  tools: {
    swc: {
      jsc: {
        transform: {
          react: {
            runtime: 'automatic',
          },
        },
      },
    },
    rspack(config) {
      config.module ??= {};
      config.module.rules ??= [];
      config.module.rules.push(
        {
          test: /\.(css|scss|less)$/,
          loader: path.resolve(import.meta.dirname, 'src/testutils/loaders/styleLoader.cjs'),
        },
        {
          test: /\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$/,
          loader: path.resolve(import.meta.dirname, 'src/testutils/loaders/fileLoader.cjs'),
        },
      );
      return config;
    },
  },
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.stories.{js,jsx,ts,tsx}',
  ],
  coverage: {
    enabled: false,
    provider: 'istanbul',
    reportsDirectory: 'coverage',
    reporter: ['lcov'],
    exclude: ['node_modules'],
  },
});
