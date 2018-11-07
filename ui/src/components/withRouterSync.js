import React, { Component } from "react"
import { withRouter } from "react-router-dom"
import getNextQuery from "../utils/getNextQuery"
import queryObjectToSearchString from "../utils/queryObjectToSearchString"
import searchStringToQueryObject from "../utils/searchStringToQueryObject"
import validateUpdateOptions from "../utils/validateUpdateOptions"

export default function withRouterSync(UnwrappedComponent) {
  class WrappedComponent extends Component {
    constructor(props) {
      super(props)
      this.updateQuery = this.updateQuery.bind(this)
    }
    updateQuery(updateOpts) {
      if (Array.isArray(updateOpts)) {
        this.updateQueryArray(updateOpts)
      } else {
        this.updateQuerySingle(updateOpts)
      }
    }
    updateQuerySingle(updateOpts) {
      if (!validateUpdateOptions(updateOpts)) {
        return console.error(`The options passed to updateQuery are invalid.`)
      }
      const { key, value, index, updateType, replace } = updateOpts
      const { history } = this.props
      const currQuery = searchStringToQueryObject(history.location.search)
      const nextQuery = getNextQuery({
        currQuery,
        key,
        value,
        index,
        updateType,
      })
      const nextSearchString = queryObjectToSearchString(nextQuery)

      if (!!replace) {
        history.replace({ search: nextSearchString })
      } else {
        history.push({ search: nextSearchString })
      }
    }
    updateQueryArray(updateArray) {
      const { history } = this.props
      let currQuery = searchStringToQueryObject(history.location.search)
      let nextQuery

      for (let i = 0; i < updateArray.length; i++) {
        if (!validateUpdateOptions(updateArray[i])) {
          return console.error(`The options passed to updateQuery are invalid.`)
          break
        }
        const { key, value, index, updateType, replace } = updateArray[i]
        nextQuery = getNextQuery({
          currQuery,
          key,
          value,
          index,
          updateType,
        })
        currQuery = nextQuery
      }

      const nextSearchString = queryObjectToSearchString(nextQuery)
      const replace = updateArray[0].replace

      if (!!replace) {
        history.replace({ search: nextSearchString })
      } else {
        history.push({ search: nextSearchString })
      }
    }
    render() {
      const { match, history } = this.props
      const pathname = history.location.pathname
      const search = history.location.search
      const query = searchStringToQueryObject(search)

      return (
        <UnwrappedComponent
          updateQuery={this.updateQuery}
          pathname={pathname}
          search={search}
          query={query}
          {...this.props}
        />
      )
    }
  }
  WrappedComponent.displayName = `withRouterSync(${UnwrappedComponent.displayName ||
    "UnwrappedComponent"})`

  return withRouter(WrappedComponent)
}
