import React, { Component } from "react"
import PropTypes from "prop-types"
import withQueryParams from "react-router-query-params"
import DebounceInput from "react-debounce-input"
import Select from "react-select"
import { get, has } from "lodash"
import Field from "../styled/Field"
import {
  stringToSelectOpt,
  selectOptToString,
  selectTheme,
  selectStyles,
} from "../../utils/reactSelectHelpers"

export const asyncDataTableFilterTypes = {
  INPUT: "INPUT",
  SELECT: "SELECT",
  CUSTOM: "CUSTOM",
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
    if (selected === null) {
      this.updateQuery(null)
      return
    }

    if (has(selected, "value")) {
      this.updateQuery(selected.value)
      return
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
    const {
      filterKey,
      queryParams,
      type,
      displayName,
      options,
      description,
    } = this.props

    const value = get(queryParams, filterKey, "")

    switch (type) {
      case asyncDataTableFilterTypes.SELECT:
        return (
          <Field label={displayName} description={description}>
            <Select
              onChange={this.handleSelectChange}
              options={options}
              value={stringToSelectOpt(value)}
              theme={selectTheme}
              styles={selectStyles}
              isClearable
            />
          </Field>
        )
      case asyncDataTableFilterTypes.INPUT:
      default:
        return (
          <Field label={displayName} description={description}>
            <DebounceInput
              className="pl-input"
              debounceTimeout={250}
              minLength={1}
              onChange={this.handleInputChange}
              type="text"
              value={value}
            />
          </Field>
        )
    }
  }
}

AsyncDataTableFilter.displayName = "AsyncDataTableFilter"

AsyncDataTableFilter.propTypes = {
  description: PropTypes.string,
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
  type: PropTypes.oneOf(Object.values(asyncDataTableFilterTypes)).isRequired,
}

AsyncDataTableFilter.defaultProps = {}

export default withQueryParams()(AsyncDataTableFilter)
