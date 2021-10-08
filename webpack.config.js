const webpack = require('webpack');
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ModuleNotFoundPlugin = require("react-dev-utils/ModuleNotFoundPlugin");
const fs = require('fs');
const InterpolateHtmlPlugin = require("react-dev-utils/InterpolateHtmlPlugin");
const errorOverlayMiddleware = require('react-dev-utils/errorOverlayMiddleware');
const evalSourceMapMiddleware = require('react-dev-utils/evalSourceMapMiddleware');
const noopServiceWorkerMiddleware = require('react-dev-utils/noopServiceWorkerMiddleware');
const redirectServedPath = require('react-dev-utils/redirectServedPathMiddleware');
const yaml = require('js-yaml');

const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = relativePath => path.resolve(appDirectory, relativePath);

const GetBuildConfigYaml = (environment = 'local') => {
  const fileName = `config.${ environment }.yaml`

  console.log(`Reading: ${ fileName }`);

  let configuration = {};

  try {
    configuration = {
      ...configuration,
      ...yaml.load(fs.readFileSync(fileName, 'utf8')),
    };
  } catch (err) {
    // Do nothing
    console.error(err);
  }

  return configuration;
}

module.exports = (env, argv) => {
  if (!env.PUBLIC_URL) {
    env.PUBLIC_URL = ''
  }

  const config = {
    target: 'web',
    entry: [
      'react-hot-loader/patch',
      './ui/index.js'
    ],
    output: {
      path: path.resolve(__dirname, 'pkg/ui/static'),
      filename: `[name].${process.env.RELEASE_REVISION || '[chunkhash]'}.js`
    },
    module: {
      rules: [
        {
          test: /\.(js|jsx)$/,
          use: 'babel-loader',
          exclude: /node_modules/
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
          loader: 'ts-loader',
          exclude: /node_modules/
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
                name: 'images/[hash]-[name].[ext]'
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
        }
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
      hot: "only",
      host: '0.0.0.0',
      port: 30000,
      webSocketServer: 'ws',
      client: {
        webSocketURL:  'wss://app.monetr.mini/ws',
      },
      onBeforeSetupMiddleware: function (devServer) {
        // Keep `evalSourceMapMiddleware` and `errorOverlayMiddleware`
        // middlewares before `redirectServedPath` otherwise will not have any effect
        // This lets us fetch source contents from webpack for the error overlay
        devServer.app.use(evalSourceMapMiddleware(devServer.server));
        // This lets us open files from the runtime error overlay.
        devServer.app.use(errorOverlayMiddleware());
      },
      onAfterSetupMiddleware: function (devServer) {
        // Redirect to `PUBLIC_URL` or `homepage` from `package.json` if url not match
        devServer.app.use(redirectServedPath("/"));

        // This service worker file is effectively a 'no-op' that will reset any
        // previous service worker registered for the same host:port combination.
        // We do this in development to avoid hitting the production cache if
        // it used the same host and port.
        // https://github.com/facebook/create-react-app/issues/2272#issuecomment-302832432
        devServer.app.use(noopServiceWorkerMiddleware("/"));
      },
    },
    plugins: [
      new webpack.DefinePlugin({
        CONFIG: JSON.stringify({}),
        RELEASE_REVISION: JSON.stringify(process.env.RELEASE_REVISION),
      }),
      new HtmlWebpackPlugin({
        inject: true,
        appMountId: 'app',
        filename: 'index.html',
        template: 'public/index.html',
        publicPath: "/",
      }),
      new webpack.ContextReplacementPlugin(/moment[\/\\]locale$/, /en/),
      // Makes some environment variables available in index.html.
      // The public URL is available as %PUBLIC_URL% in index.html, e.g.:
      // <link rel="icon" href="%PUBLIC_URL%/favicon.ico">
      // It will be an empty string unless you specify "homepage"
      // in `package.json`, in which case it will be the pathname of that URL.
      new InterpolateHtmlPlugin(HtmlWebpackPlugin, env),
      new ModuleNotFoundPlugin(resolveApp('.')),
      // new webpack.optimize.ModuleConcatenationPlugin(),
      // I'm stupid and don't know how to make this better. So just uncomment this when you need it.
      // new WebpackBundleAnalyzer(),
    ],
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
    config.output.filename = '[name].[hash].js';
  }

  return config;
};
