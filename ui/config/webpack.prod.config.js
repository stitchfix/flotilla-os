import webpack from "webpack"
import HtmlWebpackPlugin from "html-webpack-plugin"
import baseConfig from "./webpack.base.config"

module.exports = opts => {
  return {
    ...baseConfig(opts),
    plugins: [
      ...baseConfig(opts).plugins,
      new webpack.DefinePlugin({
        "process.env": {
          NODE_ENV: JSON.stringify("production"),
        },
      }),
      new HtmlWebpackPlugin({
        template: "src/index.html",
        inject: "body",
        filename: "index.html",
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
