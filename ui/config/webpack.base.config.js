import path from "path"
import webpack from "webpack"
import ExtractTextPlugin from "extract-text-webpack-plugin"

module.exports = opts => {
  const { ROOT, VENDOR } = opts
  return {
    context: ROOT,
    entry: {
      main: ["babel-polyfill", path.resolve(ROOT, "src/index.js")],
      vendor: VENDOR,
    },
    output: {
      path: path.resolve(ROOT, "build/"),
      filename: "[name].[hash].js",
      publicPath: "/",
    },
    devServer: {
      historyApiFallback: true,
    },
    module: {
      rules: [
        {
          test: /\.s?css$/,
          exclude: /node_modules/,
          use: ExtractTextPlugin.extract({
            // Including style-loader in the list of loaders will
            // result in an error (`window is not defined`).
            fallback: "style-loader",
            use: ["css-loader", "autoprefixer-loader", "sass-loader"],
          }),
        },
        {
          test: /\.less$/,
          // exclude: /node_modules/,
          loaders: ["css-loader", "less-loader"],
        },
        {
          test: /\.jsx?$/,
          exclude: /node_modules/,
          loaders: ["babel-loader"],
        },
        {
          test: /\.jpe?g$|\.gif$|\.png$|\.ttf$|\.eot$|\.svg$/,
          use: "file-loader?name=[name].[ext]?[hash]",
        },
        {
          test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
          loader: "url-loader?limit=10000&mimetype=application/fontwoff",
        },
      ],
    },
    plugins: [
      new webpack.DefinePlugin({
        "process.env": {
          FLOTILLA_API: JSON.stringify(process.env.FLOTILLA_API),
          DOCKER_REPOSITORY_HOST: JSON.stringify(
            process.env.DOCKER_REPOSITORY_HOST
          ),
          DEFAULT_CLUSTER: JSON.stringify(process.env.DEFAULT_CLUSTER),
        },
      }),
      new ExtractTextPlugin({
        filename: "[name].[hash].css",
      }),
    ],
    resolve: {
      extensions: [".js", ".jsx"],
      modules: [path.resolve(ROOT, "src"), "node_modules"],
    },
  }
}
