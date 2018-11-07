import React, { Component } from "react"
import PropTypes from "prop-types"
import { has, omit } from "lodash"
import querystring from "qs"

const offsetToPage = (offset, limit) => +offset / +limit + 1
const pageToOffset = (page, limit) => (+page - 1) * +limit

export default LIMIT => Unwrapped => {
  if (LIMIT === undefined || typeof LIMIT !== "number") {
    console.error(
      `You must in a valid "LIMIT" number for withQueryOffsetTransform`
    )
    return false
  }
  return class Wrapped extends Component {
    static displayName = `withQueryOffsetTransform(${Unwrapped.displayName ||
      "Unwrapped"})`
    static propTypes = {
      query: PropTypes.object.isRequired,
      search: PropTypes.string.isRequired,
      updateQuery: PropTypes.func.isRequired,
    }
    constructor(props) {
      super(props)
      this.updateQuery = this.updateQuery.bind(this)
      this.getQuery = this.getQuery.bind(this)
    }
    updateQuery(opts) {
      if (opts.key === "offset") {
        opts.key = "page"
        opts.value = offsetToPage(opts.value, LIMIT)
      }
      this.props.updateQuery(opts)
    }
    getQuery() {
      const { query } = this.props

      if (has(query, "page")) {
        return {
          ...omit(query, ["page"]),
          offset: pageToOffset(query.page, LIMIT),
          limit: LIMIT,
        }
      } else {
        return query
      }
    }
    render() {
      const transformedQuery = this.getQuery()
      const transformedSearchString = `?${querystring.stringify(
        transformedQuery
      )}`
      return (
        <Unwrapped
          {...this.props}
          updateQuery={this.updateQuery}
          query={transformedQuery}
          search={transformedSearchString}
        />
      )
    }
  }
}
