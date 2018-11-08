import React from "react"
import { asField } from "informed"
import Field from "./Field"

const FieldTextarea = asField(({ fieldState, fieldApi, ...props }) => (
  <Field label={props.label}>
    <textarea
      type="text"
      className="pl-textarea"
      value={fieldState.value || ""}
      onChange={evt => {
        fieldApi.setValue(evt.target.value)
      }}
    />
  </Field>
))

export default FieldTextarea
