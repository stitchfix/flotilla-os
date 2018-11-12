const OPTIONS = {
  ROOT: __dirname,
  NODE_ENV: process.env.NODE_ENV,
  VENDOR: [
    "ansi-to-react": "ansi-to-react",
    "axios": "axios",
    "immutability-helper": "immutability-helper",
    "lodash": "lodash",
    "moment": "moment",
    "prop-types": "prop-types",
    "qs": "qs",
    "react": "react",
    "react-addons-css-transition-group": "react-addons-css-transition-group",
    "react-debounce-input": "react-debounce-input",
    "react-dom": "react-dom",
    "react-feather": "react-feather",
    "react-form": "react-form",
    "react-helmet": "react-helmet",
    "react-json-view": "react-json-view",
    "react-page-visibility": "react-page-visibility",
    "react-redux": "react-redux",
    "react-router-dom": "react-router-dom",
    "react-router-query-params": "react-router-query-params",
    "react-select": "react-select",
    "redux": "redux",
    "redux-logger": "redux-logger",
    "redux-thunk": "redux-thunk",
    "styled-components": "styled-components",
    "url-join": "url-join",
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
