var webpack = require('webpack');
var path = require('path');
var config = require('./build-config');

// plugins
var HtmlWebpackPlugin = require('html-webpack-plugin');
var buildConfigPlugin = new webpack.DefinePlugin({
  "ENV": JSON.stringify(process.env.NODE_ENV),
  "BUILD_CONFIG": JSON.stringify(config[process.env.NODE_ENV]),
  // Fix: To use production build, if not dev.
  'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV === 'development' ? 'development' : 'production')
});

var devEnv = process.env.NODE_ENV === "development";

module.exports = {
  context: path.join(__dirname, "src"),
  devtool: devEnv ? "inline-sourcemap" : false,
  entry: "./index.js",
  module: {
    loaders: [
      {
        test: /\.jsx?$/,
        exclude: /(node_modules|bower_components)/,
        loader: 'babel-loader',
        query: {
          presets: ['react', 'es2015', 'stage-0'],
          plugins: ['react-html-attrs', 'transform-class-properties', 'transform-decorators-legacy'],
        }
      },
      {
          test: /\.css$/,
          loader: 'style-loader!css-loader'
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
      },
    ]
  },
  output: {
    path: __dirname + "/dist/",
    filename: "index.min.js"
  },
  plugins: devEnv ? [buildConfigPlugin] : [
    buildConfigPlugin, 
    new webpack.optimize.DedupePlugin(),
    new webpack.optimize.OccurrenceOrderPlugin(),
    new webpack.optimize.UglifyJsPlugin({
      compress: {
        warnings: false,
        conditionals: true,
        unused: true,
        comparisons: true,
        sequences: true,
        dead_code: true,
        evaluate: true,
        if_return: true,
        join_vars: true
      },
      output: {
        comments: false
      }
    }),
    new webpack.optimize.AggressiveMergingPlugin(),
    new HtmlWebpackPlugin({
      template: './index.template.html'
    })
  ],
};
