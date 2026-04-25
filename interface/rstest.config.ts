import { withRsbuildConfig } from '@rstest/adapter-rsbuild';
import { defineConfig } from '@rstest/core';

import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  extends: withRsbuildConfig({
    cwd: dirname,
    modifyRsbuildConfig(config) {
      // Strip production/dev-server settings not needed for tests
      delete config.dev;
      delete config.server;
      delete config.html;
      delete config.performance;
      delete config.security;
      // Remove PWA plugin (only added in production builds anyway, but be explicit).
      // TODO Find a way to make this less hacky?
      config.plugins = config.plugins?.filter(p => p && (p as { name: string }).name !== 'rsbuild-plugin-pwa');
      // React DOM checks `typeof IS_REACT_ACT_ENVIRONMENT !== 'undefined'` as a bare identifier. Define it at compile
      // time so the check resolves correctly inside Rspack's bundled function scope.
      config.source ??= {};
      config.source.define ??= {};
      config.source.define.IS_REACT_ACT_ENVIRONMENT = 'true';
      return config;
    },
  }),
  // Override adapter default of 'happy-dom' — must use jsdom because billing.spec.tsx relies on jsdom's internal
  // symbol-based window.location
  testEnvironment: 'jsdom',
  // Cap workers to match the PROCESSORS 4 contract set on the `interface` ctest entry. Without this, rstest defaults
  // to os.cpus().length and oversubscribes the CPU on local 8-core dev boxes where ctest only allocated 4 slots.
  pool: {
    maxWorkers: 4,
  },
  globals: true,
  setupFiles: ['./src/setupTests.ts'],
  include: ['src/**/*.spec.{ts,tsx}'],
  exclude: ['**/node_modules/**'],
  resetMocks: false,
  coverage: {
    enabled: false,
    include: ['src/**/*.{js,jsx,ts,tsx}'],
    exclude: ['**/*.d.ts', '**/*.stories.{js,jsx,ts,tsx}', '**/node_modules/**'],
    reporters: ['lcov'],
    reportsDirectory: process.env.RSTEST_COVERAGE_DIR || './coverage',
  },
});
