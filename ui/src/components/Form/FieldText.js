import React from "react"
import PropTypes from "prop-types"
import { Field as RFField } from "react-form"
import DebounceInput from "react-debounce-input"
import { get } from "lodash"
import Field from "../styled/Field"

const FieldText = props => {
  return (
    <RFField field={props.field}>
      {fieldAPI => {
        let sharedProps = {
          value: get(fieldAPI, "value", ""),
          onChange: evt => {
            fieldAPI.setValue(evt.target.value)
          },
        }

        if (props.inputRef) {
          sharedProps.ref = props.inputRef
        }

        let input

        if (props.isTextArea) {
          input = <textarea className="pl-textarea" {...sharedProps} />
        } else if (props.shouldDebounce) {
          input = (
            <DebounceInput
              {...sharedProps}
              className="pl-input"
              debounceTimeout={250}
              minLength={1}
              type={props.isNumber ? "number" : "text"}
            />
          )
        } else {
          input = (
            <input
              type={props.isNumber ? "number" : "text"}
              className="pl-input"
              {...sharedProps}
            />
          )
        }

        return (
          <Field
            label={props.label}
            isRequired={props.isRequired}
            description={props.description}
            error={fieldAPI.error}
          >
            {input}
          </Field>
        )
      }}
    </RFField>
  )
}

FieldText.propTypes = {
  description: PropTypes.string,
  field: PropTypes.string.isRequired,
  inputRef: PropTypes.func,
  isNumber: PropTypes.bool.isRequired,
  isRequired: PropTypes.bool.isRequired,
  isTextArea: PropTypes.bool.isRequired,
  label: PropTypes.string,
  shouldDebounce: PropTypes.bool.isRequired,
}

FieldText.defaultProps = {
  isNumber: false,
  isRequired: false,
  isTextArea: false,
  shouldDebounce: false,
}

export default FieldText
