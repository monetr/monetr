const webpack = require('webpack');
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ModuleNotFoundPlugin = require("react-dev-utils/ModuleNotFoundPlugin");
const fs = require('fs');
const InterpolateHtmlPlugin = require("react-dev-utils/InterpolateHtmlPlugin");

const appDirectory = fs.realpathSync(process.cwd());
const resolveApp = relativePath => path.resolve(appDirectory, relativePath);

module.exports = (env, argv) => {
  if (!env.PUBLIC_URL) {
    env.PUBLIC_URL = ''
  }

  const config = {
    entry: [
      'react-hot-loader/patch',
      './src/index.js'
    ],
    output: {
      path: path.resolve(__dirname, 'build'),
      filename: '[name].[contenthash].js'
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
          test: /\.svg$/,
          use: 'file-loader'
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
        '.ts'
      ],
      alias: {
        'react-dom': '@hot-loader/react-dom'
      },
      modules: [path.resolve(__dirname, 'src'), 'node_modules'],
    },
    devServer: {
      contentBase: './public',
      historyApiFallback: true,
    },
    plugins: [
      new HtmlWebpackPlugin({
        inject: true,
        appMountId: 'app',
        filename: 'index.html',
        template: 'public/index.html'
      }),
      new webpack.ContextReplacementPlugin(/moment[\/\\]locale$/, /en/),
      // Makes some environment variables available in index.html.
      // The public URL is available as %PUBLIC_URL% in index.html, e.g.:
      // <link rel="icon" href="%PUBLIC_URL%/favicon.ico">
      // It will be an empty string unless you specify "homepage"
      // in `package.json`, in which case it will be the pathname of that URL.
      new InterpolateHtmlPlugin(HtmlWebpackPlugin, env),
      new ModuleNotFoundPlugin(resolveApp('.')),
    ],
    optimization: {
      runtimeChunk: 'single',
      splitChunks: {
        cacheGroups: {
          vendor: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            chunks: 'all'
          }
        }
      }
    }
  };

  if (argv.hot) {
    // Cannot use 'contenthash' when hot reloading is enabled.
    config.output.filename = '[name].[hash].js';
  }

  return config;
};
