import path from 'path';
import { defineConfig } from 'vitest/config';

const interfaceSource = path.resolve(__dirname, 'src');

export default defineConfig({
  root: interfaceSource,
  css: {
    postcss: {
      plugins: [],
    },
  },
  resolve: {
    alias: {
      '@monetr/interface': interfaceSource,
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: path.join(interfaceSource, 'setupTests.ts'),
  },
});
