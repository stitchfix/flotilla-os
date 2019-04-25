const path = require("path")
const Dotenv = require("dotenv-webpack")

module.exports = opts => {
  const { ROOT, VENDOR } = opts
  return {
    context: ROOT,
    entry: {
      main: path.resolve(ROOT, "src/index.tsx"),
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
        { test: /\.tsx?$/, loader: "awesome-typescript-loader" },
        {
          enforce: "pre",
          test: /\.js$/,
          loader: "source-map-loader",

          // Prevents a source-map-loader warning from being logged per:
          // https://github.com/angular/angular-cli/issues/7115
          exclude: [path.join(process.cwd(), "node_modules")],
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
      extensions: [".ts", ".tsx", ".js", ".json"],
      modules: [path.resolve(ROOT, "src"), "node_modules"],
    },
  }
}
