import React, { Component } from "react"
import PropTypes from "prop-types"
import { NestedField } from "react-form"
import { get } from "lodash"
import FieldText from "./FieldText"

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
    const { field, values } = this.props

    return (
      <div>
        <button onClick={this.handleAddClick}>add</button>
        {!!values &&
          values.map((v, i) => (
            <NestedField
              key={`${field}-${i}`}
              field={[field, i]}
              component={props => {
                return (
                  <div
                    style={{
                      display: "flex",
                      flexFlow: "row nowrap",
                      justifyContent: "flex-start",
                      alignItems: "flex-end",
                    }}
                  >
                    <FieldText field="key" label="Key" isRequired />
                    <FieldText field="value" label="Value" isRequired />
                    <button onClick={this.handleRemoveClick}>X</button>
                  </div>
                )
              }}
            />
          ))}
      </div>
    )
  }
}

FieldKeyValue.displayName = "FieldKeyValue"

FieldKeyValue.propTypes = {
  addValue: PropTypes.func.isRequired,
  field: PropTypes.string.isRequired,
  removeValue: PropTypes.func.isRequired,
  values: PropTypes.array.isRequired,
}

export default FieldKeyValue
