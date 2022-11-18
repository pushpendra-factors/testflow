const path = require('path');

module.exports = {
  webpackDirAlias: {
    factorsComponents: path.resolve(
      __dirname,
      './src/components/factorsComponents'
    ),
    Components: path.resolve(__dirname, './src/components'),
    svgIcons: path.resolve(__dirname, './src/components/svgIcons'),
    Reducers: path.resolve(__dirname, './src/reducers'),
    Utils: path.resolve(__dirname, './src/utils'),
    Styles: path.resolve(__dirname, './src/styles'),
    hooks: path.resolve(__dirname, './src/hooks'),
    Views: path.resolve(__dirname, './src/Views'),
    Attribution: path.resolve(__dirname, './src/features/attribution'),
    Context: path.resolve(__dirname, './src/contexts')
  },
  eslintDirAlias: {
    map: [
      ['Utils', './src/utils'],
      ['Components', './src/components'],
      ['factorsComponents', './src/components/factorsComponents'],
      ['Reducers', './src/reducers'],
      ['svgIcons', './src/components/svgIcons'],
      ['Styles', './src/styles'],
      ['hooks', './src/hooks'],
      ['Views', './src/Views'],
      ['Attribution', './src/features/attribution'],
      ['Context', './src/contexts']
    ],
    extensions: ['.ts', '.tsx', '.js', '.jsx', '.json']
  }
};
