import React from "react"
import PropTypes from "prop-types"
import { Field } from "redux-form"

const asField = UnwrappedComponent => {
  const WrappedComponent = props => {
    const { name, onChange, fieldProps, ...custom } = props

    return (
      <Field
        name={name}
        onChange={onChange}
        component={UnwrappedComponent}
        props={{ custom }}
        {...fieldProps}
      />
    )
  }

  WrappedComponent.displayName = `asField(${UnwrappedComponent.displayName ||
    "UnwrappedComponent"})`
  WrappedComponent.propTypes = {
    fieldProps: PropTypes.object,
    name: PropTypes.string.isRequired,
    onChange: PropTypes.func,
  }

  return WrappedComponent
}

export default asField
