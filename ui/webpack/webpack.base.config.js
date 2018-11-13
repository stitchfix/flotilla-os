const path = require("path")
const webpack = require("webpack")
const Dotenv = require("dotenv-webpack")

module.exports = opts => {
  const { ROOT, VENDOR } = opts
  return {
    context: ROOT,
    entry: {
      main: path.resolve(ROOT, "src/index.js"),
      vendor: VENDOR,
    },
    output: {
      path: path.resolve(ROOT, "build", "static"),
      filename: "[name].[hash].js",
      publicPath: "/",
    },
    devServer: {
      historyApiFallback: true,
    },
    module: {
      rules: [
        {
          test: /\.js$/,
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
    plugins: [new Dotenv()],
    resolve: {
      extensions: [".js", ".jsx"],
      modules: [path.resolve(ROOT, "src"), "node_modules"],
    },
  }
}
