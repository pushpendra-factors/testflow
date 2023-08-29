var path = require('path');

module.exports = {
  devServer: {
    static: {
      directory: path.join(__dirname, '.'),
    },
  },
  mode: "development",
  watch: true
};