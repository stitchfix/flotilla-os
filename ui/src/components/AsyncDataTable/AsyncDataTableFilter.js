import React from "react"
import PropTypes from "prop-types"
import FieldText from "../Form/FieldText"
import FieldSelect from "../Form/FieldSelect"

export const asyncDataTableFilterTypes = {
  INPUT: "INPUT",
  SELECT: "SELECT",
  CUSTOM: "CUSTOM",
}

const AsyncDataTableFilter = props => {
  const { field, type, displayName, description } = props
  const sharedProps = {
    label: displayName,
    field,
    description,
  }

  switch (type) {
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

export default AsyncDataTableFilter
