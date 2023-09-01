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
    Context: path.resolve(__dirname, './src/contexts'),
    Constants: path.resolve(__dirname, './src/constants'),
    Routes: path.resolve(__dirname, './src/routes'),
    HOC: path.resolve(__dirname, './src/HOC'),
    Onboarding: path.resolve(__dirname, './src/features/onboarding')
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
      ['Context', './src/contexts'],
      ['Constants', './src/constants'],
      ['Routes', './src/routes'],
      ['HOC', './src/HOC'],
      ['Onboarding', './src/features/onboarding']
    ],
    extensions: ['.ts', '.tsx', '.js', '.jsx', '.json']
  },
  jestDirAlias: {
    '^Utils(.*)$': '<rootDir>/src/utils$1',
    '^Components(.*)$': '<rootDir>/src/components$1',
    '^factorsComponents(.*)$': '<rootDir>/src/components/factorsComponents$1',
    '^Reducers(.*)$': '<rootDir>/src/reducers$1',
    '^svgIcons(.*)$': '<rootDir>/src/components/svgIcons$1',
    '^Styles(.*)$': '<rootDir>/src/styles$1',
    '^hooks(.*)$': '<rootDir>/src/hooks$1',
    '^Views(.*)$': '<rootDir>/src/Views$1',
    '^Attribution(.*)$': '<rootDir>/src/features/attribution$1',
    '^Context(.*)$': '<rootDir>/src/contexts$1',
    '^Constants(.*)$': '<rootDir>/src/Constants$1',
    '^Routes(.*)$': '<rootDir>/src/routes$1',
    '^HOC(.*)$': '<rootDir>/src/HOC$1',
    '^Onboarding(.*)$': '<rootDir>/src/features/onboarding$1'
  }
};
