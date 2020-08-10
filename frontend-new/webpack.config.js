var webpack = require('webpack');
var path = require('path');
var config = require('./build-config');

// plugins
const HtmlWebPackPlugin = require("html-webpack-plugin");
var CopyWebpackPlugin = require('copy-webpack-plugin');

var buildConfigPlugin = new webpack.DefinePlugin({
  "ENV": JSON.stringify(process.env.NODE_ENV),
  "BUILD_CONFIG": JSON.stringify(config[process.env.NODE_ENV]),
  // Fix: To use production build, if not dev.
  'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV === 'development' ? 'development' : 'production')
});

const HtmlPlugin = new HtmlWebPackPlugin({
  filename: "./src/index.html" 
});

var isDev = process.env.NODE_ENV === "development";
var isStaging = process.env.NODE_ENV === "staging";

function getBuildPath() {
  return __dirname + "/dist/" + process.env.NODE_ENV;
}

module.exports = {
  entry: './src/index.js',
  devtool: (isDev || isStaging) ? "inline-sourcemap" : false,
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: ['babel-loader']
      },
      {
        test: /\.(css|sass|scss)$/,
        use: [
          { loader: 'style-loader' },
          { loader: 'css-loader' },
          { loader: 'sass-loader' }
        ]
      },
      {
        test: /\.(eot|woff|woff2|ttf|svg|png|jpg|jpeg|gif)(\?\S*)?$/,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 100000,
              name: '[name].[ext]',
            },
          },
        ],
      }
    ]
  },
  plugins: [
    buildConfigPlugin,
    HtmlPlugin,
    new CopyWebpackPlugin([{ from: './src/assets', to: 'assets' }]),
  ],
  resolve: {
    extensions: ['*', '.js', '.jsx']
  },
  output: {
    path: getBuildPath(),
    publicPath: '/',
    filename: 'index.min.js'
  },
  devServer: {
    historyApiFallback: true,
  }
};
