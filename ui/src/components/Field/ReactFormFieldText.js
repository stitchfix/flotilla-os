import React from "react"
import { Field } from "react-form"
import { get } from "lodash"
import FieldText from "./FieldText"

const ReactFormFieldText = props => (
  <Field field={props.field}>
    {fieldAPI => (
      <FieldText
        {...props}
        value={get(fieldAPI, "value", "")}
        onChange={value => fieldAPI.setValue(value)}
        error={get(fieldAPI, "error", null)}
      />
    )}
  </Field>
)

export default ReactFormFieldText
