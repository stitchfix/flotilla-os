import * as React from "react"
import { FieldArray, FastField } from "formik"
import { Button, FormGroup, Classes, Intent, Icon } from "@blueprintjs/core"
import { Env } from "../types"
import { IconNames } from "@blueprintjs/icons"
import { envFieldSpec } from "../constants"

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
      <div className={Classes.LABEL}>{envFieldSpec.label}</div>
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
          <FormGroup label={i === 0 ? "Name" : null}>
            <FastField
              name={`${envFieldSpec.name}[${i}].name`}
              className={Classes.INPUT}
            />
          </FormGroup>
          <FormGroup label={i === 0 ? "Value" : null}>
            <FastField
              name={`${envFieldSpec.name}[${i}].value`}
              className={Classes.INPUT}
            />
          </FormGroup>
          <Button
            onClick={() => {
              remove(i)
            }}
            type="button"
            intent={Intent.DANGER}
            style={i === 0 ? { transform: `translateY(8px)` } : {}}
            icon={IconNames.CROSS}
          ></Button>
        </div>
      ))}
    </div>
  </div>
)

const ConnectedEnvFieldArray: React.FunctionComponent<{}> = () => (
  <FieldArray name={envFieldSpec.name}>
    {({ form, push, remove }) => (
      <EnvFieldArray values={form.values.env} push={push} remove={remove} />
    )}
  </FieldArray>
)

export default ConnectedEnvFieldArray
