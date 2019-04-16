module.exports = {
  entry: "./src/index.js",
  output: {
    path: __dirname + "/npm_package/",
    filename: "index.js",
    library: "factors",
    libraryTarget: "umd",
    umdNamedDefine: true
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