import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import {
  withRouterSync,
  Loader,
  queryUpdateTypes,
  withStateFetch,
} from "aa-ui-components"
import querystring from "query-string"
import { has, isEmpty, isEqual } from "lodash"
import SortHeader from "./SortHeader"
import withQueryOffsetTransform from "./withQueryOffsetTransform"

const validateOpts = opts => {
  if (!has(opts, "limit") || typeof opts.limit !== "number") {
    console.error("You must pass a `limit` option (number) to withServerList.")
  }

  if (!has(opts, "getUrl") || typeof opts.getUrl !== "function") {
    console.error(
      "You must pass a `getUrl` option (function) to withServerList."
    )
  }
}

export default function withServerList(opts = {}) {
  return UnwrappedComponent => {
    class WrappedComponent extends Component {
      static displayName = `withServerList(${UnwrappedComponent.displayName ||
        "UnwrappedComponent"})`
      constructor(props) {
        super(props)
        this.fetch = this.fetch.bind(this)
      }
      componentWillMount() {
        validateOpts(opts)
      }
      componentDidMount() {
        const { query } = this.props

        if (isEmpty(query) && has(opts, "defaultQuery")) {
          this.props.updateQuery(
            Object.keys(opts.defaultQuery).map(key => ({
              key,
              value: opts.defaultQuery[key],
              updateType: queryUpdateTypes.SHALLOW,
              replace: true,
            }))
          )
        } else {
          this.fetch(
            opts.getUrl(this.props, { ...this.props.query, limit: opts.limit })
          )
        }
      }
      componentWillReceiveProps(nextProps) {
        if (!isEqual(this.props.query, nextProps.query)) {
          this.fetch(
            opts.getUrl(this.props, { ...nextProps.query, limit: opts.limit })
          )
        }
      }
      fetch(url) {
        this.props.fetch(url)
      }
      render() {
        return <UnwrappedComponent {...this.props} fetch={this.fetch} />
      }
    }

    // Note: withoutHOCStack is used for testing.
    return {
      withHOCStack: withRouterSync(
        withQueryOffsetTransform(opts.limit)(withStateFetch(WrappedComponent))
      ),
      withoutHOCStack: WrappedComponent,
    }
  }
}
