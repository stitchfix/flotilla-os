const path = require("path")
const HtmlWebpackPlugin = require("html-webpack-plugin")
const baseConfig = require("./webpack.base.config")

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
      new HtmlWebpackPlugin({
        template: "src/index.html",
        inject: "body",
        filename: path.resolve(opts.ROOT, "build", "index.html"),
        appMountId: "root",
        favicon: "src/assets/favicon.png",
      }),
    ],
  }
}
