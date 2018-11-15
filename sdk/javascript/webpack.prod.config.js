module.exports = {
  entry: "./src/index.js",
  output: {
    filename: "factors.prod.js",
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