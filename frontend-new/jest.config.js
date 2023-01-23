const { jestDirAlias } = require('./dirAlias');
var config = require('./build-config');

module.exports = {
  collectCoverage: false,
  collectCoverageFrom: ['src/**/*.{js,jsx}'],
  coverageDirectory: 'coverage',
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  moduleNameMapper: {
    ...jestDirAlias,
    '\\.(css|less|scss)$': '<rootDir>/__mocks__/styleMock.js',
    '\\.(eot|woff|woff2|ttf|svg|png|jpg|jpeg|gif)$':
      '<rootDir>/__mocks__/fileMock.js'
  },
  globals: {
    BUILD_CONFIG: config[process.env.NODE_ENV || 'development']
  }
};
