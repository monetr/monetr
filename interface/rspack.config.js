const path = require('path');
const ReactRefreshPlugin = require('@rspack/plugin-react-refresh');
const rspack = require('@rspack/core');

module.exports = (env, _argv) => {
  const envName = process.env.NODE_ENV ?? 'development';
  const isDevelopment = envName !== 'production';
  console.log(`environment: ${envName}`);

  if (!env.PUBLIC_URL) {
    env.PUBLIC_URL = '';
  }

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

  // HMR replacement gets **fucked** if we are using content hash. So use name when we are
  // in development mode.
  let filename = isDevelopment ? '[name]' : '[contenthash]';

  // Make it so that the websocket still works if we are running yarn start normally.
  const wsProto = process.env.WS_PROTO || 'ws';
  let websocketUrl = `${wsProto}://${domainName}/ws`;

  // This is used for GitPod and CodeSpaces editor environments. Allowing hot reloading when working in the cloud.
  if (process.env.CLOUD_MAGIC === 'magic' && process.env.MONETR_UI_DOMAIN_NAME) {
    websocketUrl = `${wsProto}://${domainName}/ws`;
  }

  /** @type {import('@rspack/cli').Configuration} */
  const rspackConfig = {
    mode: isDevelopment ? 'development' : 'production',
    target: 'web',
    entry: './src/index.tsx',
    experiments: {
      css: true,
    },
    output:{
      publicPath: '/',
      path: path.resolve(__dirname, '../server/ui/static'),
      filename: `assets/scripts/${filename}.js`,
      cssFilename: `assets/styles/${filename}.css`,
      cssChunkFilename: `assets/styles/${filename}.css`,
    },
    devtool: isDevelopment ? 'inline-source-map' : 'source-map',
    devServer: {
      allowedHosts: 'all',
      static: {
        directory: 'public',
      },
      historyApiFallback: true,
      host: developmentLite ? 'localhost' : '0.0.0.0',
      port: 3000,
      webSocketServer: 'ws',
      liveReload: true,
      client: {
        webSocketTransport: 'ws',
        webSocketURL: developmentLite ? undefined : websocketUrl,
        progress: true,
      },
      proxy: developmentLite ? [
        { // When we are in development-lite mode, proxy API calls to the upstream API server that they have specified.
          context: ['/api'],
          target: `https://${developmentLiteTarget}`,
          changeOrigin: true,
          cookieDomainRewrite: 'localhost',
          ws: true, // For file uploads
        }
      ] : undefined,
    },
    resolve: {
      preferRelative: false,
      extensions: [
        '.js',
        '.jsx',
        '.tsx',
        '.ts',
        '.svg',
        '.scss',
        '.css',
      ],
      modules: [
        // This makes the absolute imports work properly.
        path.resolve(__dirname, 'src'),
        'node_modules',
      ],
      alias: {
        '@monetr/interface': path.resolve(__dirname, 'src'),
      },
    },
    optimization: {
      runtimeChunk: true,
      splitChunks: {
        chunks: 'all',
        minSize: 1000,
        minChunks: 1,
        maxAsyncRequests: 30,
        maxInitialRequests: 30,
        cacheGroups: {
          defaultVendors: {
            test: /[\\/]node_modules[\\/]/,
            priority: -10,
            minChunks: 2,
          },
          default: {
            minChunks: 2,
            priority: -20,
            reuseExistingChunk: true,
          },
        },
      },
    },
    module: {
      rules: [
        {
          test: /\.(js|jsx)$/,
          use: {
            loader: 'builtin:swc-loader',
            options: {
              jsc: {
                parser: {
                  syntax: 'ecmascript',
                  jsx: true,
                },
                transform: {
                  react: {
                    development: isDevelopment,
                    refresh: isDevelopment,
                  },
                },
              },
            },
          },
        },
        {
          test: /\.(ts|tsx)$/,
          use: {
            loader: 'builtin:swc-loader',
            options: {
              jsc: {
                parser: {
                  syntax: 'typescript',
                  tsx: true,
                },
                transform: {
                  react: {
                    development: isDevelopment,
                    refresh: isDevelopment,
                  },
                },
              },
            },
          },
          type: 'javascript/auto',
        },
        {
          test: /\.(sass|scss)$/,
          use: [
            {
              loader: 'sass-loader',
              options: {
                sassOptions: {
                  quietDeps: true,
                },
              },
            },
            {
              loader: 'postcss-loader',
            },
          ],
          type: 'css/auto',
        },
        {
          test: /\.css$/,
          use: [
            {
              loader: 'postcss-loader',
            },
          ],
          type: 'css/auto',
        },
        {
          test: /\.(woff|woff2|eot|ttf)$/,
          type: 'asset',
          parser: {
            dataUrlCondition: {
              maxSize: 8 * 1024,
            },
          },
          generator: {
            filename: 'assets/font/[contenthash][ext][query]',
          },
        },
        {
          test: /\.svg$/,
          type: 'asset',
          parser: {
            dataUrlCondition: {
              maxSize: 8 * 1024, // 8KB
            },
          },
          generator: {
            filename: 'assets/img/[contenthash][ext][query]',
          },
        },
        {
          test: /\.(png|jpe?g|gif)$/,
          type: 'asset',
          parser: {
            dataUrlCondition: {
              maxSize: 8 * 1024, 
            },
          },
          generator: {
            filename: 'assets/img/[chashontenthash][ext][query]',
          },
        },
      ],
    },
    plugins: [
      isDevelopment && new ReactRefreshPlugin(),
      new rspack.DefinePlugin({
        CONFIG: JSON.stringify({}),
        REVISION: JSON.stringify(process.env.RELEASE_REVISION),
        RELEASE: JSON.stringify(process.env.RELEASE_VERSION),
        NODE_VERSION: process.version,
      }),
      new rspack.CopyRspackPlugin({
        patterns: [
          {
            from: 'public/manifest.json',
            to: 'manifest.json',
          },
          {
            from: 'public/logo192.png',
            to: 'logo192.png',
          },
          {
            from: 'public/logo512.png',
            to: 'logo512.png',
          },
          {
            from: 'public/logo192transparent.png',
            to: 'logo192transparent.png',
          },
          {
            from: 'public/logo512transparent.png',
            to: 'logo512transparent.png',
          },
          {
            from: 'public/robots.txt',
            to: 'robots.txt',
          },
        ],
      }),
      new rspack.HtmlRspackPlugin({
        minify: true,
        template: 'public/index.html',
        filename: 'index.html',
        favicon: 'public/favicon.ico',
        templateParameters: {
          // When we are doing local dev then don't use anything, maybe use an env var in the future but thats it. But
          // for a production build add the go template string in so that the server can provide the DSN.
          SENTRY_DSN: isDevelopment ? '' : '{{ .SentryDSN }}',
        },
      }),
    ].filter(Boolean),
  };

  return rspackConfig;
};
