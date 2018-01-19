import webpack from "webpack"
import HtmlWebpackPlugin from "html-webpack-plugin"
import baseConfig from "./webpack.base.config"

module.exports = opts => {
  return {
    ...baseConfig(opts),
    plugins: [
      ...baseConfig(opts).plugins,
      new HtmlWebpackPlugin({
        template: "src/index.html",
        inject: "body",
        filename: "index.html",
        appMountId: "root",
        favicon: "src/assets/favicon.png",
      }),
    ],
  }
}
