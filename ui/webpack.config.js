const OPTIONS = {
  ROOT: __dirname,
  NODE_ENV: process.env.NODE_ENV,
  VENDOR: [
    "ansi-to-react",
    "axios",
    "lodash",
    "moment",
    "qs",
    "react",
    "react-debounce-input",
    "react-dom",
    "react-feather",
    "react-helmet",
    "react-json-view",
    "react-resize-detector",
    "react-router-dom",
    "react-select",
    "react-window",
    "styled-components",
    "url-join",
  ],
}

module.exports = (() => {
  switch (process.env.NODE_ENV) {
    case "production":
      return require("./webpack/webpack.prod.config.js")
    case "dev":
      return require("./webpack/webpack.dev.config.js")
    default:
      return require("./webpack/webpack.dev.config.js")
  }
})()(OPTIONS)
