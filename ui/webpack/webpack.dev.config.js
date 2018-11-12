const HtmlWebpackPlugin = require("html-webpack-plugin")
const baseConfig = require("./webpack.base.config")

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
