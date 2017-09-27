import React, { Component } from 'react'
import { withRouter } from 'react-router'
import { isEqual } from 'lodash'
import queryString from 'query-string'
import { checkStatus } from '../utils/'
import { TablePageButton } from './'

const orderOpts = {
  ASC: 'asc',
  DESC: 'desc'
}

const DEFAULT_QUERY = {
  page: 1,
  order: orderOpts.ASC,
}

export default function serverTableConnect(opts) {
  const urlRoot = opts.urlRoot
  const limit = 20
  const initialQuery = opts.initialQuery ? {
    ...DEFAULT_QUERY,
    ...opts.initialQuery,
  } : DEFAULT_QUERY

  return (UnwrappedComponent) => {
    const WrappedComponent = class extends Component {
      constructor(props) {
        super(props)
        this.handleQueryChange = this.handleQueryChange.bind(this)
        this.handleSortChange = this.handleSortChange.bind(this)
      }
      state = {
        isFetching: false,
        data: {},
        total: 100
      }
      componentDidMount() {
        const query = {
          ...initialQuery,
          ...this.props.location.query
        }
        this.replaceUrlQuery(query)
        this.fetchData(query)
      }
      componentDidUpdate(prevProps) {
        const { location } = this.props
        if (!isEqual(prevProps.location.query, location.query)) {
          this.fetchData(location.query)
        }
      }
      fetchData(query) {
        this.setState({ isFetching: true })
        const page = !!query.page ? query.page : 1
        const _query = Object.keys(query).reduce((acc, key) => {
          if (key !== 'page') {
            acc[key] = query[key]
          }
          return acc
        }, {
          limit,
          offset: ((page - 1) * limit)
        })
        const url = `${urlRoot(this.props)}${queryString.stringify(_query)}`
        fetch(url)
          .then(checkStatus)
          .then(res => res.json())
          .then((res) => {
            this.setState({
              isFetching: false,
              total: res.total,
              data: res // What you do to this is up to you.
            })
          })
          .catch((err) => {
            console.error(err)
          })
      }
      replaceUrlQuery(query) {
        const { router, location } = this.props
        const urlQuery = !!location.query ? location.query : {}

        router.replace({
          pathname: location.pathname,
          query: {
            ...urlQuery,
            ...query
          }
        })
      }
      handleQueryChange(key, value) {
        this.replaceUrlQuery({
          [key]: value,
          page: 1
        })
      }
      handleSortChange(sortBy) {
        const { location } = this.props
        let order
        if (location.query.sort_by === sortBy) {
          order = location.query.order === orderOpts.DESC ?
            orderOpts.ASC :
            orderOpts.DESC
        } else {
          order = orderOpts.ASC
        }

        this.replaceUrlQuery({
          sort_by: sortBy,
          order,
          page: 1
        })
      }
      handlePaginationChange(page) {
        this.replaceUrlQuery({ page })
      }
      renderPageButtons() {
        const currentPage = parseInt(this.props.location.query.page, 10)
        const totalPages = Math.ceil(this.state.total / limit)
        const noOfButtons = 5
        const noOfButtonsToRender = noOfButtons > totalPages ? totalPages : noOfButtons

        return (
          <div className="flex ff-rn j-c a-c" style={{ width: '100%', marginTop: 12, marginBottom: 24 }}>
            {
              [...Array(noOfButtonsToRender).keys()].map((n) => {
                let pageNumber
                if (currentPage <= Math.ceil(noOfButtons / 2)) {
                  pageNumber = n + 1
                } else if (currentPage > Math.ceil(noOfButtons / 2) &&
                          ((currentPage + Math.floor(noOfButtons / 2)) <= totalPages)) {
                  pageNumber = (currentPage + n + 1) - Math.ceil(noOfButtons / 2)
                } else {
                  pageNumber = (totalPages + n + 1) - noOfButtons
                }

                return (
                  <TablePageButton
                    key={`table-page-button-${n}`}
                    pageNumber={pageNumber}
                    onClick={(page) => { this.handlePaginationChange(page) }}
                    isActive={pageNumber === currentPage}
                  />
                )
              })
            }
          </div>
        )
      }
      render() {
        return (
          <div>
            <UnwrappedComponent
              onQueryChange={this.handleQueryChange}
              onSortChange={this.handleSortChange}
              data={this.state.data}
              query={this.props.location.query}
              isFetching={this.state.isFetching}
              forceRefetch={() => { this.fetchData(this.props.location.query) }}
              {...this.props}
            />
            {this.state.total ? this.renderPageButtons() : null}
          </div>
        )
      }
    }
    return withRouter(WrappedComponent)
  }
}
