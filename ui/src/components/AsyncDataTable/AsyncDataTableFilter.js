import React from "react"
import PropTypes from "prop-types"
import { get, isString, isArray } from "lodash"
import FieldText from "../Field/FieldText"
import FieldSelect from "../Field/FieldSelect"
import FieldKeyValue from "../Field/FieldKeyValue"
import QueryParams from "../QueryParams/QueryParams"

export const asyncDataTableFilterTypes = {
  INPUT: "INPUT",
  SELECT: "SELECT",
  CUSTOM: "CUSTOM",
  KV: "KV",
}

const AsyncDataTableFilter = props => {
  const { field, type, displayName, description, value, onChange } = props
  const sharedProps = {
    label: displayName,
    field,
    description,
    value,
    onChange,
  }

  switch (type) {
    case asyncDataTableFilterTypes.KV:
      return null
    case asyncDataTableFilterTypes.SELECT:
      return <FieldSelect {...sharedProps} {...props} />
    case asyncDataTableFilterTypes.INPUT:
    default:
      return <FieldText {...sharedProps} shouldDebounce />
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

export default props => (
  <QueryParams>
    {({ queryParams, setQueryParams }) => (
      <AsyncDataTableFilter
        {...props}
        value={get(queryParams, props.field, "")}
        onChange={value => {
          setQueryParams({
            [props.field]: value,
          })
        }}
      />
    )}
  </QueryParams>
)
