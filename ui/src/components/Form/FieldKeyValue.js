import React, { Component } from "react"
import PropTypes from "prop-types"
import { NestedField } from "react-form"
import { X } from "react-feather"
import { isEmpty } from "lodash"
import styled from "styled-components"
import FieldText from "./FieldText"
import Button from "../styled/Button"
import Field, { FieldDescription } from "../styled/Field"
import intentTypes from "../../constants/intentTypes"

const NestedKV = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: flex-start;
  width: 100%;

  & > * {
    margin-right: 6px;
    &:last-child {
      margin-right: 0;
    }

    &:nth-child(3) {
      transform: translateY(18px);
    }
  }
`

class FieldKeyValue extends Component {
  state = {
    keyValue: "",
    valueValue: "",
    isKeyInputFocused: false,
    isValueInputFocused: false,
  }

  componentDidMount() {
    window.addEventListener("keypress", this.handleKeypress)
  }

  shouldAddField = () => {
    const { isKeyRequired, isValueRequired } = this.props
    const {
      keyValue,
      valueValue,
      isKeyInputFocused,
      isValueInputFocused,
    } = this.state

    if (!isKeyInputFocused && !isValueInputFocused) {
      return false
    }

    if (isKeyRequired === true && isEmpty(keyValue)) {
      return false
    }

    if (isValueRequired === true && isEmpty(valueValue)) {
      return false
    }

    return true
  }

  addField = () => {
    const { addValue, field, keyField, valueField } = this.props
    const { keyValue, valueValue } = this.state

    addValue(field, { [keyField]: keyValue, [valueField]: valueValue })
    this.resetState()
  }

  resetState = () => {
    this.setState({ keyValue: "", valueValue: "" }, () => {
      this.keyInput.focus()
    })
  }

  handleKeypress = evt => {
    if (evt.keyCode === 13) {
      evt.preventDefault()
      evt.stopPropagation()

      if (this.shouldAddField()) {
        this.addField()
      }
    }
  }

  toggleKeyInputFocus = () => {
    console.log("A")
    this.setState(prevState => ({
      isKeyInputFocused: !prevState.isKeyInputFocused,
    }))
  }
  toggleValueInputFocus = () => {
    console.log("B")
    this.setState(prevState => ({
      isValueInputFocused: !prevState.isValueInputFocused,
    }))
  }

  handleRemoveClick = index => {
    const { removeValue, field } = this.props

    removeValue(field, index)
  }

  render() {
    const { field, values, label, keyField, valueField } = this.props

    return (
      <Field label={label}>
        {!!values &&
          values.map((v, i) => (
            <NestedField key={`${field}-${i}`} field={[field, i]}>
              <NestedKV>
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
              </NestedKV>
            </NestedField>
          ))}
        <NestedKV>
          <Field
            field={keyField}
            label="Key"
            isRequired
            description="Press enter to add."
          >
            <input
              type="text"
              className="pl-input"
              value={this.state.keyValue}
              onChange={evt => {
                this.setState({ keyValue: evt.target.value })
              }}
              ref={x => {
                this.keyInput = x
              }}
              onFocus={this.toggleKeyInputFocus}
              onBlur={this.toggleKeyInputFocus}
            />
          </Field>
          <Field field={keyField} label="Value" isRequired>
            <input
              type="text"
              className="pl-input"
              value={this.state.valueValue}
              onChange={evt => {
                this.setState({ valueValue: evt.target.value })
              }}
              onFocus={this.toggleValueInputFocus}
              onBlur={this.toggleValueInputFocus}
            />
          </Field>
        </NestedKV>
      </Field>
    )
  }
}

FieldKeyValue.displayName = "FieldKeyValue"

FieldKeyValue.propTypes = {
  addValue: PropTypes.func.isRequired,
  field: PropTypes.string.isRequired,
  isKeyRequired: PropTypes.bool.isRequired,
  isValueRequired: PropTypes.bool.isRequired,
  keyField: PropTypes.string.isRequired,
  label: PropTypes.string,
  removeValue: PropTypes.func.isRequired,
  valueField: PropTypes.string.isRequired,
  values: PropTypes.array.isRequired,
}

FieldKeyValue.defaultProps = {
  isKeyRequired: true,
  isValueRequired: false,
  keyField: "name",
  valueField: "value",
}

export default FieldKeyValue
