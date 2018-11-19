import React from "react"
import { Field } from "react-form"
import { get } from "lodash"
import FieldSelect from "./FieldSelect"

const ReactFormFieldSelect = props => (
  <Field field={props.field}>
    {fieldAPI => (
      <FieldSelect
        {...props}
        value={get(fieldAPI, "value", "")}
        onChange={value => {
          fieldAPI.setValue(value)
        }}
        error={get(fieldAPI, "error", null)}
      />
    )}
  </Field>
)

export default ReactFormFieldSelect
