module.exports = {
  entry: "./src/app.js",
  output: {
    filename: "bundle-prod.js",
    library: "factors",
    libraryTarget: "umd"
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
  mode: "production",
  watch: false
};