import React from "react"
import { asField } from "informed"
import Field from "./Field"

const FieldText = asField(({ fieldState, fieldApi, ...props }) => (
  <Field label={props.label}>
    <input
      type="text"
      className="pl-input"
      value={fieldState.value || ""}
      onChange={evt => {
        fieldApi.setValue(evt.target.value)
      }}
    />
  </Field>
))

export default FieldText
