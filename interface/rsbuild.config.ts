import { defineConfig, type RsbuildConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import path from 'path';

export default defineConfig(({ envMode }) => {
  const envName = envMode ?? 'development';
  const isDevelopment = envName === 'development';
  console.log(`environment: ${envName}`);


  // Make it so that the websocket still works if we are running yarn start normally.
  const wsProto = process.env.WS_PROTO || 'ws';
  let websocketUrl = `${wsProto}://monetr.local/ws`;

  // This is used for GitPod and CodeSpaces editor environments. Allowing hot reloading when working in the cloud.
  if (process.env.CLOUD_MAGIC === 'magic' && process.env.MONETR_UI_DOMAIN_NAME) {
    websocketUrl = `${wsProto}://${ process.env.MONETR_UI_DOMAIN_NAME }/ws`;
  }

  return {
    source: {
      entry: {
        index: './src/index.tsx',
      },
      alias: {
        '@monetr/interface': path.resolve(__dirname, 'src'),
      },
      define: {
        CONFIG: JSON.stringify({}),
        REVISION: JSON.stringify(process.env.RELEASE_REVISION),
        RELEASE: JSON.stringify(process.env.RELEASE_VERSION),
        NODE_VERSION: process.version,
      },
    },
    dev: {
      progressBar: true,
      client: {
        protocol: wsProto as 'wss' | 'ws',
        host: 'monetr.local',
        port: '443',
      },
    },
    server: {
      port: 443,
      host: '0.0.0.0',
    },
    plugins: [pluginReact()],
    html: {
      template: 'public/index.html',
      favicon: 'public/favicon.ico',
      templateParameters: {
        // When we are doing local dev then don't use anything, maybe use an env var in the future but thats it. But
        // for a production build add the go template string in so that the server can provide the DSN.
        SENTRY_DSN: isDevelopment ? '' : '{{ .SentryDSN }}',
      },
    },
    output: {
      minify: {
        jsOptions: {
          mangle: false,
        },
      },
      targets: ['web'],
      assetPrefix: '/',
      cleanDistPath: false,
      distPath: {
        root: '../server/ui/static',
        js: 'assets/scripts',
        css: 'assets/styles',
        font: 'assets/fonts',
      },
      copy: {
        patterns: [
          {
            from: 'public/logo192.png',
            to: 'logo192.png',
          },
          {
            from: 'public/manifest.json',
            to: 'manifest.json',
          },
          {
            from: 'public/logo512.png',
            to: 'logo512.png',
          },
          {
            from: 'public/robots.txt',
            to: 'robots.txt',
          },
        ],
      },
    },
  } satisfies RsbuildConfig;
});
