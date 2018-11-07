import React, { Component } from "react"
import withRouterSync from "./withRouterSync"
import queryUpdateTypes from "../utils/queryUpdateTypes"
import withStateFetch from "./withStateFetch"
import { has, isEmpty, isEqual } from "lodash"
import withQueryOffsetTransform from "./withQueryOffsetTransform"

export default function withServerList(opts = {}) {
  return UnwrappedComponent => {
    class WrappedComponent extends Component {
      static displayName = `withServerList(${UnwrappedComponent.displayName ||
        "UnwrappedComponent"})`
      constructor(props) {
        super(props)
        this.fetch = this.fetch.bind(this)
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

      componentDidUpdate(prevProps) {
        if (!isEqual(this.props.query, prevProps.query)) {
          this.fetch(
            opts.getUrl(this.props, { ...this.props.query, limit: opts.limit })
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

    return {
      withHOCStack: withRouterSync(
        withQueryOffsetTransform(opts.limit)(withStateFetch(WrappedComponent))
      ),
      withoutHOCStack: WrappedComponent,
    }
  }
}
