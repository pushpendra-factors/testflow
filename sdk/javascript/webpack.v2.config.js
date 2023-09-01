const CONFIG = {
  entry: "./src/index.js",
  output: {
    // This file can be copied with random file names for generated asset url.
    filename: "factors.v1.js", 
    library: "_faitracker",
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