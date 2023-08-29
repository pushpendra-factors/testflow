var path = require('path');

module.exports = {
  entry: "./src/index.js",
  output: {
    filename: "factors.dev.js",
    library: "factors",
    libraryTarget: "var"
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: "babel-loader"
        }
      }
    ]
  },
  mode: "development",
  watch: true
};