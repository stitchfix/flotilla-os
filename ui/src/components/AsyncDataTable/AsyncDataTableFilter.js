import React, { Component } from "react"
import PropTypes from "prop-types"
import withQueryParams from "react-router-query-params"
import DebounceInput from "react-debounce-input"
import Select from "react-select"
import { get, has } from "lodash"

import FormGroup from "../FormGroup"

const filterTypes = {
  INPUT: "INPUT",
  SELECT: "SELECT",
}

class AsyncDataTableFilter extends Component {
  constructor(props) {
    super(props)
    this.handleInputChange = this.handleInputChange.bind(this)
    this.handleSelectChange = this.handleSelectChange.bind(this)
    this.updateQuery = this.updateQuery.bind(this)
  }

  handleInputChange(evt) {
    this.updateQuery(evt.target.value)
  }

  handleSelectChange(selected) {
    if (has(selected, "value")) {
      this.updateQuery(selected.value)
    }
  }

  updateQuery(value) {
    const { setQueryParams, filterKey } = this.props

    setQueryParams({
      [filterKey]: value,
      page: 1,
    })
  }

  render() {
    const { filterKey, queryParams, type, displayName, options } = this.props
    const value = get(queryParams, filterKey, "")
    switch (type) {
      case filterTypes.SELECT:
        return (
          <FormGroup
            label={displayName}
            input={
              <Select
                onChange={this.handleSelectChange}
                options={options}
                value={value}
              />
            }
          />
        )
      case filterTypes.INPUT:
      default:
        return (
          <FormGroup
            label={displayName}
            input={
              <DebounceInput
                className="pl-input"
                debounceTimeout={250}
                minLength={1}
                onChange={this.handleInputChange}
                type="text"
                value={value}
              />
            }
          />
        )
    }
  }
}

AsyncDataTableFilter.propTypes = {
  displayName: PropTypes.string.isRequired,
  filterKey: PropTypes.string.isRequired,
  options: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string.isRequired,
      value: PropTypes.string.isRequired,
    })
  ),
  queryParams: PropTypes.object.isRequired,
  setQueryParams: PropTypes.func.isRequired,
  type: PropTypes.oneOf(Object.values(filterTypes)).isRequired,
}
AsyncDataTableFilter.defaultProps = {}

export default withQueryParams()(AsyncDataTableFilter)
