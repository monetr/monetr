const fs = require('fs');
const webpack = require('webpack');
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ReactRefreshWebpackPlugin = require('@pmmmwh/react-refresh-webpack-plugin');
const ReactRefreshTypeScript = require('react-refresh-typescript');

const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = relativePath => path.resolve(appDirectory, relativePath);

module.exports = (env, argv) => {
  const isDevelopment = !(process.env.NODE_ENV === 'production' || argv.mode === 'production');

  if (!env.PUBLIC_URL) {
    env.PUBLIC_URL = '';
  }

  let filename = `[name].${ process.env.RELEASE_REVISION || '[chunkhash]' }.js`;
  if (!isDevelopment) {
    filename = `[name].js`;
  }

  const config = {
    mode: isDevelopment ? 'development' : 'production',
    target: 'web',
    entry: !isDevelopment ? [
      './ui/index.tsx'
    ] : [
      'react-hot-loader/patch',
      './ui/index.tsx'
    ],
    output: {
      path: path.resolve(__dirname, 'pkg/ui/static'),
      filename: filename,
      sourceMapFilename: '[name].js.map'
    },
    module: {
      rules: [
        {
          test: /\.(js|jsx)$/,
          exclude: /node_modules/,
          use: [
            {
              loader: require.resolve('babel-loader'),
              options: {
                plugins: [isDevelopment && require.resolve('react-refresh/babel')].filter(Boolean),
              },
            },
          ],
        },
        {
          test: /\.css$/,
          use: [
            'style-loader',
            {
              loader: 'css-loader',
              options: {
                importLoaders: 1
              }
            },
            'postcss-loader'
          ]
        },
        {
          test: /\.ts(x)?$/,
          exclude: /node_modules/,
          use: [
            {
              loader: require.resolve('ts-loader'),
              options: {
                getCustomTransformers: () => ({
                  before: [isDevelopment && ReactRefreshTypeScript()].filter(Boolean),
                }),
                transpileOnly: isDevelopment,
              },
            },
          ],
        },
        {
          test: /\.scss$/,
          use: [
            'style-loader',
            'css-loader',
            'sass-loader'
          ]
        },
        {
          test: /\.(svg)$/,
          use: [
            {
              loader: 'file-loader',
              options: {
                name: argv.hot ? 'images/[name].[hash].[ext]' : 'images/[name].[ext]',
              }
            },
          ],
        },
        {
          test: /\.png$/,
          use: [
            {
              loader: 'url-loader',
              options: {
                mimetype: 'image/png'
              }
            }
          ]
        },
        {
          test: /\.jpe?g$/,
          use: [
            {
              loader: 'url-loader',
              options: {
                mimetype: 'image/jpeg'
              }
            }
          ]
        },
      ]
    },
    resolve: {
      extensions: [
        '.js',
        '.jsx',
        '.tsx',
        '.ts',
        '.svg'
      ],
      alias: {
        'react-dom': '@hot-loader/react-dom'
      },
      modules: [path.resolve(__dirname, 'ui'), 'node_modules'],
    },
    devtool: 'inline-source-map',
    devServer: {
      allowedHosts: 'all',
      static: {
        directory: path.resolve(__dirname, 'public')
      },
      historyApiFallback: true,
      hot: 'only',
      host: '0.0.0.0',
      port: 30000,
      webSocketServer: 'ws',
      liveReload: true,
      client: {
        webSocketTransport: 'ws',
        webSocketURL: 'wss://app.monetr.mini/ws',
        progress: true,
      },
    },
    plugins: [
      new webpack.DefinePlugin({
        CONFIG: JSON.stringify({}),
        REVISION: JSON.stringify(process.env.RELEASE_REVISION),
        RELEASE: JSON.stringify(process.env.RELEASE_VERSION),
      }),
      new HtmlWebpackPlugin({
        inject: true,
        appMountId: 'app',
        filename: 'index.html',
        template: 'public/index.html',
        publicPath: '/',
      }),
      new webpack.ContextReplacementPlugin(/moment[\/\\]locale$/, /en/),
      isDevelopment && new webpack.HotModuleReplacementPlugin(),
      isDevelopment && new ReactRefreshWebpackPlugin({
        overlay: false,
      }),
      new webpack.LoaderOptionsPlugin({
        minimize: false,
      })
    ].filter(Boolean),
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
          default: {
            minChunks: 2,
            priority: -20,
            reuseExistingChunk: true,
          },
        },
      }
    }
  };

  if (argv.hot) {
    // Cannot use 'contenthash' when hot reloading is enabled.
    config.output.filename = '[name].[fullhash].js';
  }

  return config;
};
