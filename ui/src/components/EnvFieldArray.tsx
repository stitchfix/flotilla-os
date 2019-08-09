import * as React from "react"
import { FieldArray, FastField } from "formik"
import { Button, FormGroup, Classes, Intent } from "@blueprintjs/core"
import { Env } from "../types"

export type Props = {
  values: Env[]
  push: (env: Env) => void
  remove: (index: number) => void
}

export const EnvFieldArray: React.FunctionComponent<Props> = ({
  values,
  push,
  remove,
}) => (
  <div>
    <div className="flotilla-env-field-array-header">
      <div className={Classes.LABEL}>Env</div>
      <Button
        onClick={() => {
          push({ name: "", value: "" })
        }}
        type="button"
        className="flotilla-env-field-array-add-button"
      >
        Add
      </Button>
    </div>
    <div>
      {values.map((env: Env, i: number) => (
        <div key={i} className="flotilla-env-field-array-item">
          <FormGroup label="Name">
            <FastField name={`env[${i}].name`} className={Classes.INPUT} />
          </FormGroup>
          <FormGroup label="Value">
            <FastField name={`env[${i}].value`} className={Classes.INPUT} />
          </FormGroup>
          <Button
            onClick={() => {
              remove(i)
            }}
            type="button"
            intent={Intent.DANGER}
          >
            Remove
          </Button>
        </div>
      ))}
    </div>
  </div>
)

const ConnectedEnvFieldArray: React.FunctionComponent<{}> = () => (
  <FieldArray name="env">
    {({ form, push, remove }) => (
      <EnvFieldArray values={form.values.env} push={push} remove={remove} />
    )}
  </FieldArray>
)

export default ConnectedEnvFieldArray
