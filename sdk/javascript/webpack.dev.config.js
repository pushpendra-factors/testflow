module.exports = {
  entry: "./src/app.js",
  output: {
    filename: "bundle-dev.js",
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