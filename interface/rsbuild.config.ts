/* eslint-disable no-console */
import { pluginPWA } from './pluginPWA';

import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginSass } from '@rsbuild/plugin-sass';
import path from 'path';

const envName = process.env.NODE_ENV ?? 'development';
console.log(`Building for environment: ${envName}`);
const isDevelopment = envName !== 'production';
const interfaceSource = path.resolve(__dirname, 'src');

// If we are using development lite, then this changes the behavior of the config significantly. We instead proxy the
// staging or production API here to allow for frontend development only against real data. Requires a staging or
// production account.
const developmentLite = Boolean(process.env.MONETR_DEVELOPMENT_LITE ?? false);
const developmentLiteTarget = process.env.MONETR_DEVELOPMENT_LITE_TARGET ?? 'my.monetr.dev';
if (developmentLite) {
  console.log(`development lite environment will be used, upstream: ${developmentLiteTarget}`);
}

// HMR replacement gets **fucked** if we are using content hash. So use name when we are
// in development mode.
const filename = isDevelopment ? '[name]' : '[contenthash:8]';

export default defineConfig({
  mode: isDevelopment ? 'development' : 'production',
  source: {
    alias: {
      '@monetr/interface': interfaceSource,
    },
    define: {
      CONFIG: JSON.stringify({}),
      REVISION: JSON.stringify(process.env.RELEASE_REVISION),
      RELEASE: JSON.stringify(process.env.RELEASE_VERSION),
      NODE_VERSION: process.version,
    },
  },
  dev: {
    hmr: isDevelopment,
    liveReload: isDevelopment,
  },
  server: {
    publicDir: {
      name: path.resolve(__dirname, 'public'),
      copyOnBuild: false,
    },
    port: 3000,
    historyApiFallback: true,
    host: developmentLite ? 'localhost' : '0.0.0.0',
    proxy: developmentLite ? [
      { // When we are in development-lite mode, proxy API calls to the upstream API server that they have specified.
        context: ['/api'],
        target: `https://${developmentLiteTarget}`,
        changeOrigin: true,
        cookieDomainRewrite: 'localhost',
        ws: true, // For file uploads
      },
    ] : undefined,
  },
  html: {
    template: path.resolve(__dirname, 'public/index.html'),
    templateParameters: {
      // When we are doing local dev then don't use anything, maybe use an env var in the future but thats it. But
      // for a production build add the go template string in so that the server can provide the DSN.
      SENTRY_DSN: isDevelopment ? `${process.env.MONETR_SENTRY_DSN ?? ''}` : '{{ .SentryDSN }}',
    },
    favicon: path.resolve(__dirname, '../images/favicon.ico'),
    mountId: 'root',
  },
  output: {
    target: 'web',
    distPath: {
      root: path.resolve(__dirname, '../server/ui/static'),
      js: 'assets/scripts',
      css: 'assets/styles',
      image: 'assets/images',
      font: 'assets/fonts',
    },
    filename: isDevelopment ? undefined : {
      js: `${filename}.js`,
      css: `${filename}.css`,
      image: `[name].${filename}[ext]`,
      font: `${filename}[ext]`,
    },
    cleanDistPath: false, // Handled by cmake
    charset: 'utf8',
    filenameHash: true,
    manifest: false,
    minify: {
      js: true,
      css: true,
    },
    copy: [
      {
        from: 'public/manifest.json',
        to: 'manifest.json',
      },
      {
        from: 'public/robots.txt',
        to: 'robots.txt',
      },
    ],
  },
  security: {
    sri: {
      enable: !isDevelopment,
      algorithm: 'sha512',
    },
  },
  plugins: [
    pluginReact(),
    pluginSass(),
    !isDevelopment && pluginPWA({
      logo: path.resolve(__dirname, '../images/logo.png'),
      background: '#19161f',
    }),
  ].filter(item => Boolean(item)),
});
