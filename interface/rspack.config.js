const path = require('path');

module.exports = (env, _argv) => {
  const envName = Object.keys(env).pop() ?? process.env.NODE_ENV;
  const isDevelopment = envName === 'development';
  console.log(`environment: ${envName}`);

  if (!env.PUBLIC_URL) {
    env.PUBLIC_URL = '';
  }

  // HMR replacement gets **fucked** if we are using content hash. So use name when we are
  // in development mode.
  let filename = isDevelopment ? '[name]' : '[contenthash]';

  // Make it so that the websocket still works if we are running yarn start normally.
  const wsProto = process.env.WS_PROTO || 'ws';
  let websocketUrl = `${wsProto}://monetr.local/ws`;

  // This is used for GitPod and CodeSpaces editor environments. Allowing hot reloading when working in the cloud.
  if (process.env.CLOUD_MAGIC === 'magic' && process.env.MONETR_UI_DOMAIN_NAME) {
    websocketUrl = `${wsProto}://${ process.env.MONETR_UI_DOMAIN_NAME }/ws`;
  }

  const config = {
    experiments: {
      incrementalRebuild: true,
    },
    builtins: {
      react: {
        runtime: 'automatic',
        development: isDevelopment,
        refresh: isDevelopment,
      },
      presetEnv: {
        coreJs: '3',
      },
      define: {
        CONFIG: JSON.stringify({}),
        REVISION: JSON.stringify(process.env.RELEASE_REVISION),
        RELEASE: JSON.stringify(process.env.RELEASE_VERSION),
        NODE_VERSION: process.version,
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
      html: [
        {
          template: 'public/index.html',
          filename: 'index.html',
          favicon: 'public/favicon.ico',
          templateParameters: {
            // When we are doing local dev then don't use anything, maybe use an env var in the future but thats it. But
            // for a production build add the go template string in so that the server can provide the DSN.
            SENTRY_DSN: isDevelopment ? '' : '{{ .SentryDSN }}',
          },
        },
      ],
    },
    mode: isDevelopment ? 'development' : 'production',
    target: 'web',
    entry: './src/index.tsx',
    output: {
      publicPath: '/',
      path: '../server/ui/static',
      filename: `assets/scripts/${filename}.js`,
      cssFilename: `assets/styles/${filename}.css`,
      cssChunkFilename: `assets/styles/${filename}.css`,
    },
    resolve: {
      preferRelative: false,
      extensions: [
        '.js',
        '.jsx',
        '.tsx',
        '.ts',
        '.svg',
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
    devtool: isDevelopment ? 'inline-source-map' : 'source-map',
    devServer: {
      allowedHosts: 'all',
      static: {
        directory: 'public',
      },
      historyApiFallback: true,
      host: '0.0.0.0',
      port: 30000,
      webSocketServer: 'ws',
      liveReload: true,
      client: {
        webSocketTransport: 'ws',
        webSocketURL: websocketUrl,
        progress: true,
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
          test: /\.(ts |jsx?)$/,
          type: 'jsx',
          exclude: /node_modules/,
        },
        {
          test: /\.scss$/,
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
          type: 'css',
        },
        {
          test: /\.css$/,
          use: [
            {
              loader: 'postcss-loader',
            },
          ],
          type: 'css',
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
              maxSize: 1 * 1024 * 1024, // 1MB
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
  };

  return config;
};
