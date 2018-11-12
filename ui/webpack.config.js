const OPTIONS = {
  ROOT: __dirname,
  NODE_ENV: process.env.NODE_ENV,
  VENDOR: [
    "ansi-to-react",
    "axios",
    "immutability-helper",
    "lodash",
    "moment",
    "prop-types",
    "qs",
    "react",
    "react-addons-css-transition-group",
    "react-debounce-input",
    "react-dom",
    "react-feather",
    "react-form",
    "react-helmet",
    "react-json-view",
    "react-page-visibility",
    "react-router-dom",
    "react-router-query-params",
    "react-select",
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
