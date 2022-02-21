var webpack = require('webpack');
var path = require('path');
var config = require('./build-config');

const getCSSModuleLocalIdent = require('react-dev-utils/getCSSModuleLocalIdent');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;
const cssRegex = /\.css$/;
const cssModuleRegex = /\.module\.css$/;
const sassRegex = /\.(scss|sass)$/;
const sassModuleRegex = /\.module\.(scss|sass)$/;
const analyzerArg = JSON.parse(process.env.npm_config_argv);
const analyzer = analyzerArg.original.includes('--analyze');

const getStyleLoaders = (cssOptions, preProcessor) => {
  const loaders = [
    require.resolve('style-loader'),
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
        ident: 'postcss',
        plugins: () => [
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
  ];
  if (preProcessor) {
    loaders.push(require.resolve(preProcessor));
  }
  return loaders;
};

// plugins
const HtmlWebPackPlugin = require('html-webpack-plugin');
var CopyWebpackPlugin = require('copy-webpack-plugin');

var buildConfigPlugin = new webpack.DefinePlugin({
  ENV: JSON.stringify(process.env.NODE_ENV),
  BUILD_CONFIG: JSON.stringify(config[process.env.NODE_ENV]),
  // Fix: To use production build, if not dev.
  'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV === 'development' ? 'development' : 'production')
});

const HtmlPlugin = new HtmlWebPackPlugin({
  template: './src/index.template.html',
  filename: './index.html',
  title: 'Caching',
});

var isDev = process.env.NODE_ENV === 'development';
var isStaging = process.env.NODE_ENV === 'staging';

function getBuildPath() {
  return __dirname + '/dist/' + process.env.NODE_ENV;
}

module.exports = {
  entry: './src/index.js',
  devtool: (isDev || isStaging) ? 'inline-sourcemap' : false, // default false
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: ['babel-loader']
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
    extensions: ['*', '.js', '.jsx'],
    alias: {
      factorsComponents: path.resolve(__dirname, './src/components/factorsComponents'),
      Components: path.resolve(__dirname, './src/components'),
      svgIcons: path.resolve(__dirname, './src/components/svgIcons'),
      Reducers: path.resolve(__dirname, './src/reducers'),
      Utils: path.resolve(__dirname, './src/utils'),
      Styles: path.resolve(__dirname, './src/styles'),
      hooks: path.resolve(__dirname, './src/hooks'),
    }
  },
  plugins: [
    buildConfigPlugin,
    HtmlPlugin,
    new CopyWebpackPlugin([{ from: './src/assets', to: 'assets' }]),
    new BundleAnalyzerPlugin({
      analyzerMode: analyzer ? 'server' : 'disabled',
    }),
  ],
  output: {
    path: getBuildPath(),
    publicPath: '/',
    filename: '[name].[hash].js',
  },
  devServer: {
    historyApiFallback: true,
    hot: true
  }
};
