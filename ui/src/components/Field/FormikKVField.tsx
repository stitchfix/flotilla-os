import * as React from "react"
import { X } from "react-feather"
import { pick, get, isEmpty } from "lodash"
import { Field, FieldArray } from "formik"
import { Input } from "../styled/Inputs"
import Button from "../styled/Button"
import NestedKeyValueRow from "../styled/NestedKeyValueRow"
import KVFieldInput from "./KVFieldInput"
import KVFieldContainer from "./KVFieldContainer"
import { intents } from "../../.."

interface IFormikKVFieldProps {
  description?: string
  isKeyRequired: boolean
  isValueRequired: boolean
  keyField: string
  label: string
  name: string
  setFieldValue: (field: string, value: any) => void
  value: any[]
  valueField: string
}

class FormikKVField extends React.Component<IFormikKVFieldProps> {
  static defaultProps: Partial<IFormikKVFieldProps> = {
    isKeyRequired: true,
    isValueRequired: false,
    keyField: "name",
    valueField: "value",
  }

  render() {
    const {
      name,
      value,
      keyField,
      valueField,
      setFieldValue,
      label,
      description,
    } = this.props
    return (
      <FieldArray name={name}>
        {arrayHelpers => {
          return (
            <KVFieldContainer label={label} description={description}>
              {!isEmpty(value) &&
                value.map((v, i) => (
                  <NestedKeyValueRow key={i}>
                    <Field
                      name={`${name}[${i}].${keyField}`}
                      value={v[keyField]}
                      onChange={(evt: React.SyntheticEvent) => {
                        const target = evt.target as HTMLInputElement
                        setFieldValue(`${name}[${i}].${keyField}`, target.value)
                      }}
                      component={Input}
                    />
                    <Field
                      name={`${name}[${i}].${valueField}`}
                      value={v[valueField]}
                      onChange={(evt: React.SyntheticEvent) => {
                        const target = evt.target as HTMLInputElement
                        setFieldValue(
                          `${name}[${i}].${valueField}`,
                          target.value
                        )
                      }}
                      component={Input}
                    />
                    <Button
                      intent={intents.ERROR}
                      onClick={() => {
                        arrayHelpers.remove(i)
                      }}
                      type="button"
                    >
                      <X size={14} />
                    </Button>
                  </NestedKeyValueRow>
                ))}
              <KVFieldInput
                addValue={(name: string, value: any) => {
                  arrayHelpers.push({
                    [keyField]: get(value, keyField, ""),
                    [valueField]: get(value, valueField, ""),
                  })
                }}
                {...pick(this.props, [
                  "name",
                  "isKeyRequired",
                  "isValueRequired",
                  "keyField",
                  "valueField",
                ])}
              />
            </KVFieldContainer>
          )
        }}
      </FieldArray>
    )
  }
}

export default FormikKVField
