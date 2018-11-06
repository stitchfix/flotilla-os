const OPTIONS = {
  ROOT: __dirname,
  NODE_ENV: process.env.NODE_ENV,
  VENDOR: [
    "ansi-to-react",
    "axios",
    "classnames",
    "immutability-helper",
    "lodash",
    "moment",
    "qs",
    "react",
    "react-debounce-input",
    "react-dom",
    "react-feather",
    "react-helmet",
    "react-json-view",
    "react-redux",
    "react-router",
    "react-router-dom",
    "react-select",
    "react-tooltip",
    "redux",
    "redux-form",
    "redux-thunk",
  ],
}

module.exports = (() => {
  switch (process.env.NODE_ENV) {
    case "production":
      return require("./config/webpack.prod.config.js")
    case "dev":
      return require("./config/webpack.dev.config.js")
    default:
      return require("./config/webpack.dev.config.js")
  }
})()(OPTIONS)
