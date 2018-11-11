import React from "react"
import PropTypes from "prop-types"
import { Field as RFField } from "react-form"
import { get } from "lodash"
import Field from "../styled/Field"

const FieldText = props => {
  return (
    <RFField field={props.field}>
      {fieldAPI => {
        const sharedProps = {
          value: get(fieldAPI, "value", ""),
          onChange: evt => {
            fieldAPI.setValue(evt.target.value)
          },
        }

        let input

        if (props.isTextArea) {
          input = <textarea className="pl-textarea" {...sharedProps} />
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
  isNumber: PropTypes.bool.isRequired,
  isRequired: PropTypes.bool.isRequired,
  isTextArea: PropTypes.bool.isRequired,
  label: PropTypes.string,
}

FieldText.defaultProps = {
  isNumber: false,
  isRequired: false,
  isTextArea: false,
}

export default FieldText
