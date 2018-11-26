var debug = process.env.NODE_ENV !== "production";
var webpack = require('webpack');
var path = require('path');
var config = require('./buildConfig');

var buildConfigPlugin = new webpack.DefinePlugin({
  "BUILD_CONFIG": JSON.stringify(config[process.env.NODE_ENV])
});

module.exports = {
  context: path.join(__dirname, "src"),
  devtool: debug ? "inline-sourcemap" : false,
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
          test: /\.(eot|woff|woff2|ttf|svg|png|jpe?g|gif)(\?\S*)?$/,
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
    path: __dirname + "/src/",
    filename: "index.min.js"
  },
  plugins: debug ? [buildConfigPlugin] : [
    buildConfigPlugin, 
    new webpack.optimize.DedupePlugin(),
    new webpack.optimize.OccurenceOrderPlugin(),
    new webpack.optimize.UglifyJsPlugin({ mangle: false, sourcemap: false }),
  ],
};
