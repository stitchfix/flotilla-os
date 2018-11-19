import React from "react"
import PropTypes from "prop-types"
import { get, isString, isArray } from "lodash"
import { QueryParamsFieldText } from "../Field/FieldText"
import { QueryParamsFieldSelect } from "../Field/FieldSelect"
import QueryParamsKVField from "../Field/QueryParamsKVField"
import QueryParams from "../QueryParams/QueryParams"

export const asyncDataTableFilterTypes = {
  INPUT: "INPUT",
  SELECT: "SELECT",
  CUSTOM: "CUSTOM",
  KV: "KV",
}

const AsyncDataTableFilter = props => {
  const { field, type, displayName, description } = props

  const sharedProps = {
    label: displayName,
    field,
    description,
  }

  switch (type) {
    case asyncDataTableFilterTypes.KV:
      return (
        <QueryParamsKVField {...sharedProps} isKeyRequired isValueRequired />
      )
    case asyncDataTableFilterTypes.SELECT:
      return <QueryParamsFieldSelect {...sharedProps} {...props} />
    case asyncDataTableFilterTypes.INPUT:
    default:
      return <QueryParamsFieldText {...sharedProps} shouldDebounce />
  }
}

AsyncDataTableFilter.displayName = "AsyncDataTableFilter"
AsyncDataTableFilter.propTypes = {
  description: PropTypes.string,
  displayName: PropTypes.string.isRequired,
  field: PropTypes.string.isRequired,
  options: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string.isRequired,
      value: PropTypes.string.isRequired,
    })
  ),
  type: PropTypes.oneOf(Object.values(asyncDataTableFilterTypes)).isRequired,
}
AsyncDataTableFilter.defaultProps = {}

export default AsyncDataTableFilter
