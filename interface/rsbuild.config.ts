/* eslint-disable no-console */
import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginSass } from '@rsbuild/plugin-sass';
import path from 'path';

const envName = process.env.NODE_ENV ?? 'development';
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
// Domain name for local dev, only when running in docker.
const domainName = process.env.MONETR_UI_DOMAIN_NAME ?? 'monetr.local';

// Make it so that the websocket still works if we are running yarn start normally.
const wsProto = process.env.WS_PROTO || 'ws';
let websocketUrl = `${wsProto}://${domainName}/ws`;

// This is used for GitPod and CodeSpaces editor environments. Allowing hot reloading when working in the cloud.
if (process.env.CLOUD_MAGIC === 'magic' && process.env.MONETR_UI_DOMAIN_NAME) {
  websocketUrl = `${wsProto}://${domainName}/ws`;
}

export default defineConfig({
  mode: isDevelopment ? 'development' : 'production',
  source: {
    alias: {
      '@monetr/interface': interfaceSource,
    },
  },
  html: {
    template: path.resolve(__dirname, 'public/index.html'),
    templateParameters: {
      // When we are doing local dev then don't use anything, maybe use an env var in the future but thats it. But
      // for a production build add the go template string in so that the server can provide the DSN.
      SENTRY_DSN: isDevelopment ? '' : '{{ .SentryDSN }}',
    },
  },
  dev: {
    hmr: isDevelopment,
    liveReload: isDevelopment,
  },
  server: {
    publicDir: {
      name: path.resolve(__dirname, 'public'),
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
  output: {
    target: 'web',
    distPath: {
      // TODO Chunk file names with a hash or something
      root: path.resolve(__dirname, '../server/ui/static'),
      js: 'assets/scripts',
      css: 'assets/styles',
      font: 'assets/fonts',
    },
    cleanDistPath: 'auto',
    charset: 'utf8',
    filenameHash: true,
  },
  plugins: [
    pluginReact(),
    pluginSass(),
  ],
});
