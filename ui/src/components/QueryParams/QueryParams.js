import React, { Component } from "react"
import PropTypes from "prop-types"
import { withRouter } from "react-router-dom"
import { isEmpty } from "lodash"
import qs from "qs"

class QueryParams extends Component {
  getQuery = () => {
    const { search } = this.props

    if (search.length > 0) {
      return qs.parse(search.slice(1))
    }

    return {}
  }

  setQuery = (query, shouldReplace = false) => {
    const { replace, push } = this.props

    const next = qs.stringify(
      this.filterEmptyValues({
        ...this.getQuery(),
        ...query,
      }),
      { indices: false }
    )

    if (shouldReplace) {
      replace({ search: next })
    } else {
      push({ search: next })
    }
  }

  filterEmptyValues = values =>
    Object.keys(values).reduce((acc, key) => {
      if (!isEmpty(values[key])) {
        acc[key] = values[key]
      }

      return acc
    }, {})

  render() {
    return this.props.children({
      queryParams: this.getQuery(),
      setQueryParams: this.setQuery,
    })
  }
}

QueryParams.propTypes = {
  children: PropTypes.func.isRequired,
  push: PropTypes.func.isRequired,
  replace: PropTypes.func.isRequired,
  search: PropTypes.string.isRequired,
}

export default withRouter(props => (
  <QueryParams
    {...props}
    search={props.location.search}
    push={props.history.push}
    replace={props.history.replace}
  />
))
