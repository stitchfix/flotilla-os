import React, { PureComponent } from "react"
import PropTypes from "prop-types"
import DebounceInput from "react-debounce-input"
import { Field as RFField } from "react-form"
import { get, has } from "lodash"
import Field from "../styled/Field"
import { Input, Textarea } from "../styled/Inputs"
import QueryParams from "../QueryParams/QueryParams"

export const FieldText = props => {
  const {
    description,
    error,
    inputRef,
    isNumber,
    isRequired,
    isTextArea,
    label,
    onChange,
    shouldDebounce,
    value,
    field,
  } = props

  // Common props for all input components
  let sharedProps = {
    value,
    onChange: evt => {
      onChange(evt.target.value)
    },
  }

  if (!!inputRef) {
    sharedProps.ref = inputRef
  }

  // Assign input element based on various props.
  let input

  if (isTextArea) {
    input = <Textarea {...sharedProps} />
  } else if (shouldDebounce) {
    input = (
      <DebounceInput
        {...sharedProps}
        element={Input}
        debounceTimeout={250}
        minLength={1}
        type={isNumber ? "number" : "text"}
      />
    )
  } else {
    input = <Input type={isNumber ? "number" : "text"} {...sharedProps} />
  }

  return (
    <Field
      label={label}
      isRequired={isRequired}
      description={description}
      error={error}
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

export const QueryParamsFieldText = props => {
  return (
    <QueryParams>
      {({ queryParams, setQueryParams }) => (
        <FieldText
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
}

export const ReactFormFieldText = props => {
  return (
    <RFField field={props.field} validate={props.validate}>
      {fieldAPI => {
        return (
          <FieldText
            {...props}
            value={get(fieldAPI, "value", "")}
            onChange={value => fieldAPI.setValue(value)}
            error={get(fieldAPI, "error", null)}
          />
        )
      }}
    </RFField>
  )
}
