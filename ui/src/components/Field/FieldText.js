import React from "react"
import PropTypes from "prop-types"
import DebounceInput from "react-debounce-input"
import Field from "../styled/Field"
import { Input, Textarea } from "../styled/Inputs"

const FieldText = props => {
  let sharedProps = {
    value: props.value,
    onChange: evt => {
      props.onChange(evt.target.value)
    },
  }

  if (!!props.inputRef) {
    sharedProps.ref = props.inputRef
  }

  let input

  if (props.isTextArea) {
    input = <Textarea {...sharedProps} />
  } else if (props.shouldDebounce) {
    input = (
      <DebounceInput
        {...sharedProps}
        element={Input}
        debounceTimeout={250}
        minLength={1}
        type={props.isNumber ? "number" : "text"}
      />
    )
  } else {
    input = <Input type={props.isNumber ? "number" : "text"} {...sharedProps} />
  }

  return (
    <Field
      label={props.label}
      isRequired={props.isRequired}
      description={props.description}
      error={props.error}
    >
      {input}
    </Field>
  )
}

FieldText.propTypes = {
  description: PropTypes.string,
  error: PropTypes.any,
  field: PropTypes.string.isRequired,
  inputRef: PropTypes.func,
  isNumber: PropTypes.bool.isRequired,
  isRequired: PropTypes.bool.isRequired,
  isTextArea: PropTypes.bool.isRequired,
  label: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  shouldDebounce: PropTypes.bool.isRequired,
  value: PropTypes.any,
}

FieldText.defaultProps = {
  isNumber: false,
  isRequired: false,
  isTextArea: false,
  shouldDebounce: false,
}

export default FieldText
