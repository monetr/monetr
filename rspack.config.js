const path = require('path');


module.exports = (env, argv) => {
  const envName = Object.keys(env).pop() ?? process.env.NODE_ENV;

  const isDevelopment = envName === 'development';

  if (!env.PUBLIC_URL) {
    env.PUBLIC_URL = '';
  }

  let filename = `[name].${ process.env.RELEASE_REVISION || '[chunkhash]' }.js`;
  if (!isDevelopment) {
    filename = `[name].js`;
  }

  let websocketUrl = 'wss://monetr.local/ws';

  // This is used for GitPod and CodeSpaces editor environments. Allowing hot reloading when working in the cloud.
  if (process.env.CLOUD_MAGIC === 'magic' && process.env.MONETR_UI_DOMAIN_NAME) {
    websocketUrl = `wss://${ process.env.MONETR_UI_DOMAIN_NAME }/ws`;
  }
  const config = {
    builtins: {
      react: {
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
      html: [
        {
          template: 'public/index.html',
          filename: 'index.html',
          favicon: 'public/favicon.ico',
          templateParameters: {
            appMountId: 'app',
          },
        },
      ],
    },
    mode: isDevelopment ? 'development' : 'production',
    target: 'web',
    entry: './ui/index.tsx',
    output: {
      path: path.resolve(__dirname, 'pkg/ui/static'),
      filename: filename,
      // Source maps are automatically moved to $(PWD)/build/source_maps each time the UI is compiled. They will not be
      // in the path above.
      // sourceMapFilename: isDevelopment ? `[name].${ process.env.RELEASE_REVISION || '[chunkhash]' }.js.map` : '[name].[hash:8].js.map',
    },
    resolve: {
      extensions: [
        '.js',
        '.jsx',
        '.tsx',
        '.ts',
        '.svg',
      ],
      modules: [path.resolve(__dirname, 'ui'), 'node_modules'],
    },
    devtool: isDevelopment ? 'inline-source-map' : 'source-map',
    devServer: {
      allowedHosts: 'all',
      static: {
        directory: path.resolve(__dirname, 'public'),
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
      runtimeChunk: 'single',
      splitChunks: {
        chunks: 'async',
        minSize: 20000,
        minRemainingSize: 0,
        minChunks: 1,
        maxAsyncRequests: 30,
        maxInitialRequests: 30,
        enforceSizeThreshold: 50000,
        cacheGroups: {
          defaultVendors: {
            test: /[\\/]node_modules[\\/]/,
            priority: -10,
            reuseExistingChunk: true,
          },
          default:
             {
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
          test: /\.css$/,
          use: [
            {
              loader: 'postcss-loader',
            },
          ],
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
          ],
          type: 'css',
        },
        {
          test: /\.(woff|woff2|eot|ttf)$/,
          type: 'asset',
          parser: {
            dataUrlCondition: {
              maxSize: 4 * 1024,
            },
          },
          generator: {
            filename: 'assets/font/[hash][ext][query]',
          },
        },
        {
          test: /\.(png|jpe?g|gif|svg|xlsx)$/,
          type: 'asset',
          parser: {
            dataUrlCondition: {
              maxSize: 4 * 1024,
            },
          },
          generator: {
            filename: 'assets/img/[hash][ext][query]',
          },
        },
        {
          test: /\.yml$/,
          type: 'asset',
        },
        {
          test: /\.ico$/,
          type: 'asset',
          generator: {
            filename: 'assets/img/[name][ext][query]',
          },
        },
      ],
    },
  };

  if (argv.hot) {
    // Cannot use 'contenthash' when hot reloading is enabled.
    config.output.filename = '[name].[fullhash].js';
  }

  return config;
};
