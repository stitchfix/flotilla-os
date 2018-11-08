import React, { Component } from "react"
import PropTypes from "prop-types"
import { NestedField } from "react-form"
import { X } from "react-feather"
import { get } from "lodash"
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
    const { field, values, label } = this.props

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
          <Button onClick={this.handleAddClick}>Add</Button>
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
                  field="key"
                  label={i === 0 ? "Key" : null}
                  isRequired
                />
                <FieldText
                  field="value"
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
  label: PropTypes.string,
  removeValue: PropTypes.func.isRequired,
  values: PropTypes.array.isRequired,
}

export default FieldKeyValue
