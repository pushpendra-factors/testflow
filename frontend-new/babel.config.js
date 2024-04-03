module.exports = function (api) {
  api.cache(true);

  const presets = [
    '@babel/preset-env',
    '@babel/preset-react',
    '@babel/preset-typescript'
  ];
  const plugins = [
    '@babel/plugin-proposal-class-properties',
    '@babel/plugin-syntax-dynamic-import',
    '@babel/plugin-proposal-optional-chaining',
    [
      'import',
      {
        libraryName: 'antd',
        libraryDirectory: 'lib'
      },
      'antd'
    ],
    [
      'import',
      {
        libraryName: '@ant-design/icons',
        libraryDirectory: '',
        camel2DashComponentName: false
      },
      '@ant-design/icons'
    ],
    [
      'import',
      {
        libraryName: 'lodash',
        libraryDirectory: '',
        camel2DashComponentName: false
      }
    ]
  ];

  return {
    presets,
    plugins
  };
};
