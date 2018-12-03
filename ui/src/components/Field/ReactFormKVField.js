import React, { Component } from "react"
import PropTypes from "prop-types"
import { NestedField } from "react-form"
import { X } from "react-feather"
import { pick } from "lodash"
import { ReactFormFieldText } from "./FieldText"
import Button from "../styled/Button"
import Field from "../styled/Field"
import NestedKeyValueRow from "../styled/NestedKeyValueRow"
import intentTypes from "../../constants/intentTypes"
import KVFieldInput from "./KVFieldInput"
import {
  SHARED_KV_FIELD_PROPS,
  SHARED_KV_FIELD_DEFAULT_PROPS,
} from "../../utils/kvFieldHelpers"
import KVFieldContainer from "./KVFieldContainer"

export class ReactFormKVField extends Component {
  /** Removes a value specified by index. */
  handleRemoveClick = index => {
    const { removeValue, field } = this.props

    removeValue(field, index)
  }

  render() {
    const {
      field,
      label,
      values,
      keyField,
      addValue,
      valueField,
      description,
      isRequired,
      isKeyRequired,
      isValueRequired,
      validateKey,
      validateValue,
    } = this.props

    return (
      <KVFieldContainer label={label} description={description}>
        {!!values &&
          values.map((v, i) => (
            <NestedField key={`${field}-${i}`} field={[field, i]}>
              <NestedKeyValueRow>
                <ReactFormFieldText
                  field={keyField}
                  label={null}
                  isRequired={isKeyRequired}
                  validate={validateKey}
                />
                <ReactFormFieldText
                  field={valueField}
                  label={null}
                  isRequired={isValueRequired}
                  validate={validateValue}
                />
                <Button
                  intent={intentTypes.error}
                  onClick={this.handleRemoveClick.bind(this, i)}
                >
                  <X size={14} />
                </Button>
              </NestedKeyValueRow>
            </NestedField>
          ))}
        <KVFieldInput
          addValue={addValue}
          {...pick(this.props, [
            "field",
            "isKeyRequired",
            "isValueRequired",
            "keyField",
            "valueField",
          ])}
        />
      </KVFieldContainer>
    )
  }
}

ReactFormKVField.propsTypes = {
  ...SHARED_KV_FIELD_PROPS,
  addValue: PropTypes.func.isRequired,
  removeValue: PropTypes.func.isRequired,
}

ReactFormKVField.defaultProps = SHARED_KV_FIELD_DEFAULT_PROPS

export default ReactFormKVField
