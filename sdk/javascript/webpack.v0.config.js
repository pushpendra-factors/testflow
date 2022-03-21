const CONFIG = {
  entry: "./src/index.js",
  output: {
    filename: "factors.v0.js",
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
  mode: "production",
  watch: false
};

module.exports = (env) => {return CONFIG}; 