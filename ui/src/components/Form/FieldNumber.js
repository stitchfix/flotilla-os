import React from "react"
import { asField } from "informed"
import Field from "./Field"

const FieldNumber = asField(({ fieldState, fieldApi, ...props }) => (
  <Field label={props.label}>
    <input
      type="number"
      className="pl-input"
      value={fieldState.value || 0}
      onChange={evt => {
        fieldApi.setValue(evt.target.value)
      }}
    />
  </Field>
))

export default FieldNumber
