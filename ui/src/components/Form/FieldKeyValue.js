import React, { Component } from "react"
import PropTypes from "prop-types"
import { NestedField } from "react-form"
import { X } from "react-feather"
import FieldText from "./FieldText"
import Button from "../Button"
import intentTypes from "../../constants/intentTypes"

class FieldKeyValue extends Component {
  handleAddClick = () => {
    const { addValue, field } = this.props

    addValue(field, "")
  }

  handleRemoveClick = index => {
    const { removeValue, field } = this.props

    removeValue(field, index)
  }

  render() {
    const { field, values, label, keyField, valueField } = this.props

    return (
      <div>
        <div
          style={{
            display: "flex",
            flexFlow: "row nowrap",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div>{label}</div>
          <Button onClick={this.handleAddClick} type="button">
            Add
          </Button>
        </div>
        {!!values &&
          values.map((v, i) => (
            <NestedField key={`${field}-${i}`} field={[field, i]}>
              <div
                style={{
                  display: "flex",
                  flexFlow: "row nowrap",
                  justifyContent: "flex-start",
                  alignItems: "flex-end",
                }}
              >
                <FieldText
                  field={keyField}
                  label={i === 0 ? "Key" : null}
                  isRequired
                />
                <FieldText
                  field={valueField}
                  label={i === 0 ? "Value" : null}
                  isRequired
                />
                <Button
                  intent={intentTypes.error}
                  onClick={() => {
                    this.handleRemoveClick(i)
                  }}
                >
                  <X size={14} />
                </Button>
              </div>
            </NestedField>
          ))}
      </div>
    )
  }
}

FieldKeyValue.displayName = "FieldKeyValue"

FieldKeyValue.propTypes = {
  addValue: PropTypes.func.isRequired,
  field: PropTypes.string.isRequired,
  keyField: PropTypes.string.isRequired,

  label: PropTypes.string,
  removeValue: PropTypes.func.isRequired,
  valueField: PropTypes.string.isRequired,
  values: PropTypes.array.isRequired,
}

FieldKeyValue.defaultProps = {
  keyField: "name",
  valueField: "value",
}

export default FieldKeyValue
