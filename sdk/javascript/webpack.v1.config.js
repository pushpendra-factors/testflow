const CONFIG = {
    entry: "./src/index.js",
    output: {
      filename: "factors.v1.js",
      library: "factorsai",
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