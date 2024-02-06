const { eslintDirAlias } = require('./dirAlias');

module.exports = {
  env: {
    browser: true,
    es2021: true,
    jest: true
  },
  extends: [
    'plugin:react/recommended',
    'plugin:jest/recommended',
    'airbnb',
    'prettier'
  ],
  parser: '@babel/eslint-parser',
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module'
  },
  plugins: ['react', 'prettier', 'jest', 'import'],
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
    'prettier/prettier': ['error'],
    'import/extensions': [1, { json: 'ignorePackages' }],
    camelcase: 'off',
    'react/function-component-definition': [
      2,
      {
        namedComponents: [
          'function-declaration',
          'function-expression',
          'arrow-function'
        ],
        unnamedComponents: ['function-expression', 'arrow-function']
      }
    ]
  },
  settings: {
    'import/resolver': {
      node: {
        extensions: ['.js', '.jsx', '.ts', '.tsx']
      },
      alias: eslintDirAlias
    }
  },
  overrides: [
    {
      files: ['**/*.{ts,tsx}'],
      parser: '@typescript-eslint/parser',
      extends: ['plugin:@typescript-eslint/recommended'],
      rules: {
        'react/prop-types': 'off',
        'react/jsx-filename-extension': [
          2,
          {
            extensions: ['.jsx', '.tsx']
          }
        ],
        ...{
          'no-use-before-define': 'off',
          '@typescript-eslint/no-use-before-define': ['error'],
          'react/require-default-props': 'off'
        }
      }
    }
  ]
};
