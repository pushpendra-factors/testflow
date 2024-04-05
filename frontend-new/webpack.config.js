const webpack = require('webpack');
const TerserPlugin = require('terser-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');

const getCSSModuleLocalIdent = require('react-dev-utils/getCSSModuleLocalIdent');
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer');

const cssRegex = /\.css$/;
const cssModuleRegex = /\.module\.css$/;
const sassRegex = /\.(scss|sass)$/;
const sassModuleRegex = /\.module\.(scss|sass)$/;
const analyzerArg = JSON.parse(process.env.npm_config_argv);
const analyzer = analyzerArg.original.includes('--analyze');
const isDev = process.env.NODE_ENV === 'development';
const isStaging = process.env.NODE_ENV === 'staging';

const getStyleLoaders = (cssOptions, preProcessor) => {
  const loaders = [
    isDev ? require.resolve('style-loader') : MiniCssExtractPlugin.loader,
    {
      loader: require.resolve('css-loader'),
      options: cssOptions
    },
    {
      // Options for PostCSS as we reference these options twice
      // Adds vendor prefixing based on your specified browser support in
      // package.json
      loader: require.resolve('postcss-loader'),
      options: {
        // Necessary for external CSS imports to work
        // https://github.com/facebook/create-react-app/issues/2677
        postcssOptions: {
          plugins: [
            require('autoprefixer'),
            require('cssnano'),
            require('postcss-flexbugs-fixes'),
            require('postcss-preset-env')({
              autoprefixer: {
                flexbox: 'no-2009'
              },
              stage: 3
            })
          ]
        }
      }
    }
  ];
  if (preProcessor) {
    loaders.push(require.resolve(preProcessor));
  }
  return loaders;
};

// plugins
const HtmlWebPackPlugin = require('html-webpack-plugin');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const config = require('./build-config');
const { webpackDirAlias } = require('./dirAlias');

const getNodeModulesRegExp = (deps) =>
  new RegExp(`[\\/]node_modules[\\/]${deps.join('|')}`);
const reactCacheGroupDeps = [
  'react',
  'react-dom',
  'react-router',
  'react-error-boundary'
];
const reduxCacheGroupDeps = [
  'redux',
  'react-redux',
  '@reduxjs',
  'addon-redux',
  'reselect',
  'redux-promise-middleware',
  'redux-persist'
];
const chartCacheGroupDeps = ['highcharts', 'd3', 'react-pivottable'];
const thirdPartyLibsCacheGroupDeps = [
  '@sentry',
  'react-product-fruits',
  'logrocket'
];

const excludeNodeModulesRegExp = (deps) =>
  new RegExp(`[\\/]node_modules[\\/](?!(${deps.join('|')})).*`);

const vendorCacheGroupDeps = [
  ...reactCacheGroupDeps,
  ...reduxCacheGroupDeps,
  ...chartCacheGroupDeps,
  ...thirdPartyLibsCacheGroupDeps
];

const VendorCacheGroup = {
  name: 'vendor-common',
  test: excludeNodeModulesRegExp(vendorCacheGroupDeps),
  priority: -10,
  minSize: 0,
  maxSize: 3000000
};

const ReactCacheGroup = {
  name: 'vendor-react',
  test: getNodeModulesRegExp(reactCacheGroupDeps),
  priority: -1,
  minChunks: 1,
  minSize: 0,
  maxSize: 3000000
};

const ReduxCacheGroup = {
  name: 'vendor-redux',
  test: getNodeModulesRegExp(reduxCacheGroupDeps),
  priority: -2,
  minChunks: 1,
  minSize: 0,
  maxSize: 3000000
};

const ChartCacheGroup = {
  name: 'vendor-chart',
  test: getNodeModulesRegExp(chartCacheGroupDeps),
  priority: -5
};

const ThirdPartyLibsCacheGroup = {
  name: 'vendor-third-party-lib',
  test: getNodeModulesRegExp(thirdPartyLibsCacheGroupDeps),
  priority: -8,
  minChunks: 1,
  minSize: 0,
  maxSize: 3000000
};

const cacheGroups = {
  react: ReactCacheGroup,
  redux: ReduxCacheGroup,
  charts: ChartCacheGroup,
  thirdpartyLibs: ThirdPartyLibsCacheGroup,
  vendors: VendorCacheGroup
};

const splitChunks = {
  chunks: 'all',
  minChunks: 3,
  hidePathInfo: true,
  automaticNameDelimiter: '-',
  maxSize: 1000 * 1000,
  minSize: 1000 * 50,
  maxAsyncRequests: 30,
  maxInitialRequests: 30,
  cacheGroups: {
    [cacheGroups.react.name]: cacheGroups.react,
    [cacheGroups.redux.name]: cacheGroups.redux,
    [cacheGroups.charts.name]: cacheGroups.charts,
    [cacheGroups.thirdpartyLibs.name]: cacheGroups.thirdpartyLibs,
    [cacheGroups.vendors.name]: cacheGroups.vendors
  }
};

const buildConfigPlugin = new webpack.DefinePlugin({
  ENV: JSON.stringify(process.env.NODE_ENV),
  BUILD_CONFIG: JSON.stringify(config[process.env.NODE_ENV]),
  // Fix: To use production build, if not dev.
  'process.env.NODE_ENV': JSON.stringify(
    process.env.NODE_ENV === 'development' ? 'development' : 'production'
  )
});

const HtmlPlugin = new HtmlWebPackPlugin({
  template: './src/index.template.html',
  filename: './index.html',
  title: 'Caching'
});

function getBuildPath() {
  return `${__dirname}/dist/${process.env.NODE_ENV}`;
}

module.exports = {
  entry: './src/index.js',
  devtool: isDev || isStaging ? 'inline-sourcemap' : false, // default false
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: ['babel-loader']
      },
      {
        test: /\.(ts|tsx)$/,
        use: [
          'babel-loader',
          {
            loader: 'ts-loader',
            options: {
              transpileOnly: true
            }
          }
        ],
        exclude: /node_modules/
      },
      {
        test: cssRegex,
        exclude: cssModuleRegex,
        use: getStyleLoaders({
          importLoaders: 1
        })
      },
      {
        test: sassRegex,
        exclude: sassModuleRegex,
        use: getStyleLoaders({ importLoaders: 2 }, 'sass-loader')
      },
      {
        test: sassModuleRegex,
        use: getStyleLoaders(
          {
            importLoaders: 2,
            modules: true,
            getLocalIdent: getCSSModuleLocalIdent
          },
          'sass-loader'
        )
      },
      {
        test: /\.(eot|woff|woff2|ttf|svg|png|jpg|jpeg|gif)(\?\S*)?$/,
        use: [
          {
            loader: 'url-loader',
            options: {
              limit: 100000,
              name: '[name].[ext]'
            }
          }
        ]
      }
    ]
  },
  resolve: {
    extensions: ['*', '.ts', '.tsx', '.js', '.jsx'],
    alias: {
      ...webpackDirAlias
    }
  },
  plugins: [
    buildConfigPlugin,
    HtmlPlugin,
    new CopyWebpackPlugin([{ from: './src/assets', to: 'assets' }]),
    new MiniCssExtractPlugin(),
    new BundleAnalyzerPlugin({
      analyzerMode: analyzer ? 'server' : 'disabled'
    })
  ],
  optimization: {
    splitChunks,
    usedExports: true,
    removeAvailableModules: true,
    minimizer: isDev ? [] : [new TerserPlugin(), new CssMinimizerPlugin()],
    minimize: !isDev
  },
  output: {
    path: getBuildPath(),
    publicPath: '/',
    filename: '[name].[hash].js'
  },
  devServer: {
    historyApiFallback: true,
    hot: true
  }
};
