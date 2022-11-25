const { eslintDirAlias } = require('./dirAlias');

module.exports = {
  env: {
    browser: true,
    es2021: true,
    jest: true
  },
  extends: ['plugin:react/recommended', 'airbnb', 'prettier'],
  overrides: [],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module'
  },
  plugins: ['react', 'prettier'],
  rules: {
    'react/prop-types': 0,
    'react/jsx-props-no-spreading': 'off',
    'react/jsx-filename-extension': [1, { extensions: ['.js', '.jsx'] }],
    'jsx-a11y/click-events-have-key-events': 'off',
    'jsx-a11y/no-static-element-interactions': 'off',
    radix: 'off',
    'import/prefer-default-export': 'off',
    'no-nested-ternary': 'off',
    'no-plusplus': 'off',
    'prettier/prettier': ['error']
  },
  settings: {
    'import/resolver': {
      node: {
        extensions: ['.js', '.jsx', '.ts', '.tsx']
      },
      alias: eslintDirAlias
    }
  }
};
