// Configuración de Webpack 5: compila TypeScript/React y empaqueta el panel en /dist.
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');

module.exports = (_env, argv) => {
  const produccion = argv.mode === 'production';
  return {
    entry: './src/index.tsx',
    output: {
      path: path.resolve(__dirname, 'dist'),
      filename: produccion ? 'js/[name].[contenthash:8].js' : 'js/[name].js',
      publicPath: '/',
      clean: true,
    },
    resolve: { extensions: ['.tsx', '.ts', '.js'] },
    module: {
      rules: [
        { test: /\.tsx?$/, loader: 'ts-loader', exclude: /node_modules/ },
        { test: /\.css$/, use: [MiniCssExtractPlugin.loader, 'css-loader'] },
        { test: /\.(woff2?|svg)$/, type: 'asset/resource', generator: { filename: 'recursos/[name].[contenthash:8][ext]' } },
      ],
    },
    plugins: [
      new HtmlWebpackPlugin({ template: './src/index.html' }),
      new MiniCssExtractPlugin({ filename: 'css/[name].[contenthash:8].css' }),
    ],
    optimization: {
      minimizer: ['...', new CssMinimizerPlugin()],
      splitChunks: { cacheGroups: { vendor: { test: /[\\/]node_modules[\\/]/, name: 'vendor', chunks: 'all' } } },
    },
    performance: { hints: false },
    devtool: produccion ? false : 'eval-source-map',
    devServer: {
      port: 3000,
      historyApiFallback: true, // las rutas de React (/campanas/...) siempre sirven index.html
      // En desarrollo, las llamadas /api, /t y /r van al gateway (docker compose up).
      proxy: [{ context: ['/api', '/t', '/r'], target: 'http://localhost:8080' }],
    },
  };
};
