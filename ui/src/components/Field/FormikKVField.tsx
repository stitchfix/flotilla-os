import * as React from "react"
import { X } from "react-feather"
import { pick, get, isEmpty } from "lodash"
import { FastField, FieldArray } from "formik"
import Button from "../styled/Button"
import NestedKeyValueRow from "../styled/NestedKeyValueRow"
import KVFieldContainer from "./KVFieldContainer"
import { flotillaUIIntents } from "../../types"
import KVFieldInput from "./KVFieldInput"

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

class FormikKVField extends React.PureComponent<IFormikKVFieldProps> {
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
                    <FastField name={`${name}[${i}].${keyField}`} />
                    <FastField name={`${name}[${i}].${valueField}`} />
                    <Button
                      intent={flotillaUIIntents.ERROR}
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
