import path from "path"
import webpack from "webpack"
import HtmlWebpackPlugin from "html-webpack-plugin"
import baseConfig from "./webpack.base.config"

module.exports = opts => {
  const base = baseConfig(opts)
  return {
    ...base,
    output: {
      ...base.output,
      publicPath: "/static/",
    },
    plugins: [
      ...base.plugins,
      new webpack.DefinePlugin({
        "process.env": {
          NODE_ENV: JSON.stringify("production"),
        },
      }),
      new HtmlWebpackPlugin({
        template: "src/index.html",
        inject: "body",
        filename: path.resolve(opts.ROOT, "build", "index.html"),
        appMountId: "root",
        favicon: "src/assets/favicon.png",
      }),
      new webpack.optimize.CommonsChunkPlugin({
        name: "vendor",
        minChunks: Infinity,
        filename: "vendor.[hash].js",
        chunks: opts.VENDOR,
      }),
      new webpack.optimize.UglifyJsPlugin(),
    ],
  }
}
